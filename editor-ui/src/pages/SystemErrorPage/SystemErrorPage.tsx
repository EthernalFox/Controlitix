import { IconClockPause } from "@tabler/icons-react";
import { useMemo } from "react";
import { useLocation, useNavigate } from "react-router";

import { Button, Stack } from "@shared/ui";

import styles from "./SystemErrorPage.module.css";

type Variant = "403" | "404" | "500" | "session-expired";

interface SystemErrorPageProps {
  variant?: Variant;
  errorId?: string;
}

const variantMeta: Record<Variant, { code: string; from: string; to: string; title: string; description: string }> = {
  "403": {
    code: "403",
    from: "#FFA726",
    to: "#F08F0E",
    title: "Нет доступа",
    description: "У вас недостаточно прав для этого действия"
  },
  "404": {
    code: "404",
    from: "#42A5F5",
    to: "#1976D2",
    title: "Не найдено",
    description: "Страница или ресурс не найдены"
  },
  "500": {
    code: "500",
    from: "#EF5350",
    to: "#C62828",
    title: "Ошибка сервера",
    description: "Сервис временно недоступен"
  },
  "session-expired": {
    code: "",
    from: "#7E57C2",
    to: "#7E57C2",
    title: "Сессия истекла",
    description: "Выполните вход повторно"
  }
};

const resolveVariantFromPath = (pathname: string): Variant => {
  if (pathname === "/403") {
    return "403";
  }

  if (pathname === "/500") {
    return "500";
  }

  if (pathname === "/session-expired") {
    return "session-expired";
  }

  return "404";
};

export default function SystemErrorPage({ variant, errorId }: SystemErrorPageProps) {
  const navigate = useNavigate();
  const { pathname } = useLocation();

  const resolvedVariant = variant ?? resolveVariantFromPath(pathname);
  const meta = useMemo(() => variantMeta[resolvedVariant], [resolvedVariant]);

  return (
    <main className={styles.page}>
      <section className={styles.card}>
        <Stack gap="xs">
          {resolvedVariant === "session-expired" ? (
            <IconClockPause size={32} color="var(--ctrx-alarm-ack)" />
          ) : (
            <p
              className={styles.code}
              style={{
                ["--from" as string]: meta.from,
                ["--to" as string]: meta.to
              }}
            >
              {meta.code}
            </p>
          )}
          <h1 className={styles.title}>{meta.title}</h1>
          <p className={styles.description}>{meta.description}</p>
          {resolvedVariant === "500" && errorId ? <p className={styles.errorId}>id: {errorId}</p> : null}
        </Stack>

        <div className={styles.action}>
          {resolvedVariant === "500" ? (
            <Button variant="secondary" onClick={() => window.location.reload()}>
              Повторить
            </Button>
          ) : resolvedVariant === "session-expired" ? (
            <Button onClick={() => navigate("/login")}>Войти</Button>
          ) : (
            <Button variant="secondary" onClick={() => navigate("/")}>
              На главную
            </Button>
          )}
        </div>
      </section>
    </main>
  );
}
