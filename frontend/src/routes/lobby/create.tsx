import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/lobby/create')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/lobby/create"!</div>
}
