import { NavLink } from "react-router";

import { routePaths } from "@shared/libs/router";
import { Text } from "@shared/ui";

const linkStyle = ({ isActive }: { isActive: boolean }) => ({
  display: "block",
  padding: "10px 12px",
  borderRadius: 8,
  textDecoration: "none",
  color: "inherit",
  background: isActive ? "rgba(84, 116, 180, 0.18)" : "transparent"
});

export const AppNavbar = () => {
  return (
    <nav style={{ padding: 8, display: "grid", gap: 4 }}>
      <NavLink to={routePaths.objects} style={linkStyle}>
        <Text fw={500}>Объекты мониторинга</Text>
      </NavLink>
      <NavLink to={routePaths.mimics} style={linkStyle}>
        <Text fw={500}>Мнемосхемы</Text>
      </NavLink>
      <NavLink to={routePaths.devices} style={linkStyle}>
        <Text fw={500}>Устройства</Text>
      </NavLink>
    </nav>
  );
};
