import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/lobby/$code/join')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/lobby/$code/join"!</div>
}
