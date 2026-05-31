package agenttransport

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/imfeelingtheagi/netctl/internal/bus"
	"github.com/imfeelingtheagi/netctl/internal/crypto"
	agentv1 "github.com/imfeelingtheagi/netctl/internal/gen/netctl/agent/v1"
	resultv1 "github.com/imfeelingtheagi/netctl/internal/gen/netctl/result/v1"
	"github.com/imfeelingtheagi/netctl/internal/store"
	"github.com/imfeelingtheagi/netctl/internal/tenancy"
)

const heartbeatIntervalSeconds = 30

// service implements agentv1.AgentServiceServer.
type service struct {
	agentv1.UnimplementedAgentServiceServer
	pool     *pgxpool.Pool
	bus      bus.Bus
	log      *slog.Logger
	agents   store.Agents
	shutdown <-chan struct{}
	accepted atomic.Uint64 // results accepted across all StreamResults calls
}

// Register upserts the agent into its tenant's registry. The id and tenant are
// taken from the verified certificate, so this is always tenant-correct.
func (svc *service) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
	id, err := identityFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	name := req.GetHostname()
	if name == "" {
		name = id.AgentID
	}
	var agent *store.Agent
	err = tenancy.InTenant(tenancy.WithTenant(ctx, tenancy.ID(id.TenantID)), svc.pool,
		func(ctx context.Context, s tenancy.Scope) error {
			a, e := svc.agents.Register(ctx, s, id.AgentID, name, req.GetHostname(),
				req.GetAgentVersion(), id.String(), req.GetCapabilities())
			agent = a
			return e
		})
	if err != nil {
		svc.log.Error("agent register failed", "tenant", id.TenantID, "agent", id.AgentID, "error", err.Error())
		return nil, status.Error(codes.Internal, "register failed")
	}
	svc.log.Info("agent registered", "tenant", id.TenantID, "agent", id.AgentID, "hostname", req.GetHostname())
	return &agentv1.RegisterResponse{
		AgentId:                  agent.ID,
		TenantId:                 agent.TenantID,
		ConfigEpoch:              0,
		HeartbeatIntervalSeconds: heartbeatIntervalSeconds,
	}, nil
}

// Attest acknowledges the agent's identity. The mTLS handshake already proved it;
// SVID-based node/workload attestation is S-EE1.
func (svc *service) Attest(ctx context.Context, _ *agentv1.AttestRequest) (*agentv1.AttestResponse, error) {
	id, err := identityFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &agentv1.AttestResponse{Ok: true, Message: "attested " + id.String()}, nil
}

// Heartbeat marks the agent online.
func (svc *service) Heartbeat(ctx context.Context, _ *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	id, err := identityFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	err = tenancy.InTenant(tenancy.WithTenant(ctx, tenancy.ID(id.TenantID)), svc.pool,
		func(ctx context.Context, s tenancy.Scope) error {
			_, e := svc.agents.Heartbeat(ctx, s, id.AgentID)
			return e
		})
	if err != nil {
		return nil, status.Error(codes.Internal, "heartbeat failed")
	}
	return &agentv1.HeartbeatResponse{ConfigStale: false, HeartbeatIntervalSeconds: heartbeatIntervalSeconds}, nil
}

// StreamConfig pushes configuration to the agent. Placeholder: one empty epoch-0
// update, then the stream stays open until the agent disconnects. Real test/probe
// config arrives in S7+.
func (svc *service) StreamConfig(_ *agentv1.StreamConfigRequest, stream grpc.ServerStreamingServer[agentv1.StreamConfigResponse]) error {
	id, err := identityFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	if err := stream.Send(&agentv1.StreamConfigResponse{Epoch: 0}); err != nil {
		return err
	}
	svc.log.Debug("config stream opened", "tenant", id.TenantID, "agent", id.AgentID)
	select {
	case <-stream.Context().Done(): // agent disconnected
	case <-svc.shutdown: // server shutting down
	}
	return nil
}

// StreamResults accepts a stream of results, publishes each to the result bus,
// and acknowledges the count. The result's tenant + agent are taken from the
// verified certificate (never the payload), so a result is always attributed to
// the sending agent's tenant (CLAUDE.md §7 guardrails 1 and 5).
func (svc *service) StreamResults(stream grpc.ClientStreamingServer[agentv1.StreamResultsRequest, agentv1.StreamResultsResponse]) error {
	id, err := identityFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	ctx := stream.Context()
	var accepted uint64
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			svc.accepted.Add(accepted)
			return stream.SendAndClose(&agentv1.StreamResultsResponse{Accepted: accepted})
		}
		if err != nil {
			svc.accepted.Add(accepted)
			return err
		}
		if err := svc.ingest(ctx, id, req); err != nil {
			// A bus-publish failure surfaces as a stream error so the agent
			// retries the (still-buffered) batch: at-least-once delivery.
			svc.accepted.Add(accepted)
			svc.log.Error("result ingest failed", "tenant", id.TenantID, "agent", id.AgentID, "error", err.Error())
			return status.Error(codes.Unavailable, "ingest failed")
		}
		accepted++
	}
}

// ingest decodes one result, stamps the authoritative tenant + agent identity
// from the certificate, and publishes it to the result bus keyed by tenant.
func (svc *service) ingest(ctx context.Context, id crypto.SPIFFEID, req *agentv1.StreamResultsRequest) error {
	var r resultv1.Result
	if err := proto.Unmarshal(req.GetPayload(), &r); err != nil {
		// A malformed payload is a poison message: drop it (counted as accepted
		// so the agent does not wedge retrying it) rather than fail the stream.
		svc.log.Error("dropping malformed result payload", "tenant", id.TenantID, "agent", id.AgentID, "error", err.Error())
		return nil
	}
	// Authoritative identity comes from the mTLS certificate, not the payload.
	r.TenantId = id.TenantID
	r.AgentId = id.AgentID
	if svc.bus == nil {
		return nil // no bus wired (minimal server): accept and count only
	}
	value, err := proto.Marshal(&r)
	if err != nil {
		return err
	}
	return svc.bus.Publish(ctx, bus.NetworkResultsTopic, []byte(r.TenantId), value)
}
