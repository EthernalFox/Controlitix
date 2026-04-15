import { Group, Text } from "@shared/ui/components";
import { NavLink } from "react-router";

export const Logo = () => {
  return (
    <NavLink
      to={"/"}
      style={{
        textDecoration: "none"
      }}
    >
      <Group p={6}>
        <img src="/public/LightIcon.svg" alt="logo" height={48} />
        <Text size="xl">Controlitix</Text>
      </Group>
    </NavLink>
  );
};
