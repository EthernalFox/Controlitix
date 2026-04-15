import { DrawCanvas } from "@features/Draw";
import { Layout } from "@shared/ui";
import { DrawElementsPanel } from "@widgets/DrawElementsPanel";
import { DrawFooter } from "@widgets/DrawFooter";
import { DrawHeader } from "@widgets/DrawHeader";
import { DrawPropertiesPanel } from "@widgets/DrawPropertiesPanel";

export default function DrawPage() {
  return (
    <Layout
      header={<DrawHeader />}
      panel={<DrawElementsPanel />}
      aside={<DrawPropertiesPanel />}
      footer={<DrawFooter />}
      sizes={{
        headerHeight: { base: 56, sm: 64 },
        footerHeight: { base: 64, sm: 72 },
        panelWidth: { base: 240, md: 280 },
        asideWidth: { base: 280, md: 320 }
      }}
    >
      <div style={{ padding: 16, height: "100%" }}>
        <DrawCanvas />
      </div>
    </Layout>
  );
}
