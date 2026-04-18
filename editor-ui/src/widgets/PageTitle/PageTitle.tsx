import { useMatches } from "react-router";

import { Text } from "@shared/ui";

const DEFAULT_TITLE = "Controlitix";

export const PageTitle = () => {
  const matches = useMatches();
  const titleMatch = [...matches]
    .reverse()
    .find(
      (match) =>
        typeof match.handle === "object" &&
        match.handle !== null &&
        "title" in match.handle &&
        typeof match.handle.title === "string"
    );

  const title = (titleMatch?.handle as { title?: string } | undefined)?.title;

  return (
    <Text fw={600} size="lg" truncate>
      {title ?? DEFAULT_TITLE}
    </Text>
  );
};
