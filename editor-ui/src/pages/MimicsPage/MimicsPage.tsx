import { Logo } from "@shared/ui";
import { Header, Layout, Navbar } from "@shared/ui";
import { Text } from "@shared/ui";
import { AppNavbar } from "@widgets/AppNavbar";

export default function MimicsPage() {
  return (
    <Layout
      header={<Header main={<Logo />} />}
      navbar={<Navbar center={<AppNavbar />} />}
    >
      <div style={{ padding: 16 }}>
        <Text size="xl" fw={600}>
          Мнемосхемы
        </Text>
        <Text c="dimmed" mt={8}>
          Заглушка страницы. Здесь будет список мнемосхем.
        </Text>
      </div>
    </Layout>
  );
}
