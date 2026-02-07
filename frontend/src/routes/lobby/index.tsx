import { createFileRoute } from "@tanstack/react-router";
import LobbiesPage from "../../pages/LobbiesPage";

export const Route = createFileRoute("/lobby/")({
  component: LobbiesPage,
});
