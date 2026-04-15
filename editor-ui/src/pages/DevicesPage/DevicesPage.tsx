import { Logo } from "@shared/ui";
import { Header, Layout, Navbar } from "@shared/ui";
import { Text } from "@shared/ui";
import { AppNavbar } from "@widgets/AppNavbar";

export default function DevicesPage() {
  return (
    <Layout
      header={<Header main={<Logo />} />}
      navbar={<Navbar center={<AppNavbar />} />}
    >
      <div style={{ padding: 16 }}>
        <Text size="xl" fw={600}>
          Устройства
        </Text>
        <Text c="dimmed" mt={8}>
          Заглушка страницы. Здесь будет список устройств.
        </Text>
      </div>
    </Layout>
  );
}
