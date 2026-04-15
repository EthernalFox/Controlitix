import { Button, Group, Header, Logo } from "@shared/ui";

export const DrawHeader = () => {
  return (
    <Header
      before={<Logo />}
      main={
        <Group gap={8}>
          <Button variant="ghost" size="xs">
            Align
          </Button>
          <Button variant="ghost" size="xs">
            Distribute
          </Button>
          <Button variant="ghost" size="xs">
            Grid
          </Button>
          <Button variant="ghost" size="xs">
            Snap
          </Button>
        </Group>
      }
      after={
        <Group gap={8}>
          <Button variant="secondary" size="xs">
            Preview
          </Button>
          <Button variant="primary" size="xs">
            Publish
          </Button>
        </Group>
      }
    />
  );
};
