import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Providers } from '../App'
import { AppRoutes } from '../routes/AppRoutes'

/** Renders the full app (shell + routes) at a path, with a MemoryRouter. */
export function renderApp(path = '/targets') {
  return render(
    <Providers>
      <MemoryRouter initialEntries={[path]}>
        <AppRoutes />
      </MemoryRouter>
    </Providers>,
  )
}
