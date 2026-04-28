import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router";

import { useAuthStore } from "@/features/auth";
import { APP_PATHS } from "@/shared/libs/router";
import { Button, Stack, Text, TextInput } from "@/shared/ui/components";

interface RedirectState {
  from?: string;
}

const resolveRedirectPath = (from: unknown): string => {
  if (typeof from !== "string") {
    return APP_PATHS.DASHBOARD;
  }

  if (!from.startsWith("/") || from.startsWith(APP_PATHS.LOGIN)) {
    return APP_PATHS.DASHBOARD;
  }

  return from;
};

const resolveErrorMessage = (
  errorType: string | null,
  retryAfterSeconds: number
): string | null => {
  switch (errorType) {
    case "/errors/auth/invalid-credentials":
      return "Неверный логин или пароль";
    case "/errors/auth/user-disabled":
      return "Учётная запись отключена. Обратитесь к администратору";
    case "/errors/auth/rate-limited":
      if (retryAfterSeconds > 0) {
        return `Слишком много попыток. Повторите через ${retryAfterSeconds} секунд`;
      }
      return null;
    case "/errors/auth/source-disabled":
      return "Сервис аутентификации недоступен. Попробуйте позже";
    default:
      if (!errorType) {
        return null;
      }
      return "Не удалось выполнить вход. Попробуйте ещё раз";
  }
};

export const LoginForm = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const login = useAuthStore((state) => state.login);
  const status = useAuthStore((state) => state.status);
  const lastError = useAuthStore((state) => state.lastError);
  const rateLimitUntil = useAuthStore((state) => state.rateLimitUntil);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!rateLimitUntil || rateLimitUntil <= Date.now()) {
      return;
    }

    setNow(Date.now());

    const timerId = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => {
      window.clearInterval(timerId);
    };
  }, [rateLimitUntil]);

  const retryAfterSeconds = useMemo(() => {
    if (!rateLimitUntil) {
      return 0;
    }

    return Math.max(0, Math.ceil((rateLimitUntil - now) / 1000));
  }, [now, rateLimitUntil]);

  const isRateLimited = retryAfterSeconds > 0;
  const isSubmitting = status === "authenticating";
  const isDisabled = isSubmitting || isRateLimited;

  const errorMessage = resolveErrorMessage(lastError, retryAfterSeconds);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (isRateLimited) {
      return;
    }

    try {
      await login(username, password);

      const redirectState = location.state as RedirectState | null;
      const redirectPath = resolveRedirectPath(redirectState?.from);

      void navigate(redirectPath, { replace: true });
    } catch {
      // Login errors are reflected in auth-store and shown by the form.
    } finally {
      setPassword("");
    }
  };

  return (
    <form onSubmit={onSubmit}>
      <Stack gap="sm">
        <TextInput
          label="Логин"
          value={username}
          onChange={(event) => setUsername(event.currentTarget.value)}
          autoComplete="username"
          autoFocus
          required
          disabled={isDisabled}
        />
        <TextInput
          label="Пароль"
          value={password}
          onChange={(event) => setPassword(event.currentTarget.value)}
          type="password"
          autoComplete="current-password"
          required
          disabled={isDisabled}
        />

        {errorMessage ? (
          <Text size="sm" c="alarm-alarm">
            {errorMessage}
          </Text>
        ) : null}

        <Button type="submit" fullWidth loading={isSubmitting} disabled={isDisabled}>
          Войти
        </Button>
      </Stack>
    </form>
  );
};
