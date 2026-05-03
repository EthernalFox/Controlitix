import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router";

import { hasEditorRole, useAuthStore } from "@features/auth";
import { routePaths } from "@shared/libs/router";
import { Button, PasswordInput, Stack, Text, TextInput } from "@shared/ui";

interface RedirectState {
  from?: string;
}

const isAllowedRedirectTarget = (path: string): boolean => {
  if (!path.startsWith("/")) {
    return false;
  }

  if (path.startsWith(routePaths.login) || path.startsWith(routePaths.accessDenied)) {
    return false;
  }

  return true;
};

const resolveRedirectPath = (from: unknown): string => {
  if (typeof from !== "string") {
    return routePaths.objects;
  }

  return isAllowedRedirectTarget(from) ? from : routePaths.objects;
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
      return "Слишком много попыток. Повторите позже";
    case "/errors/auth/source-disabled":
      return "Сервис аутентификации недоступен. Попробуйте позже";
    case "/errors/auth/network":
      return "Не удалось связаться с сервером";
    default:
      if (!errorType) {
        return null;
      }
      return "Не удалось выполнить вход";
  }
};

export const LoginForm = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const login = useAuthStore((state) => state.login);
  const status = useAuthStore((state) => state.status);
  const user = useAuthStore((state) => state.user);
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

      const currentUser = useAuthStore.getState().user || user;
      if (!hasEditorRole(currentUser)) {
        void navigate(routePaths.accessDenied, { replace: true });
        return;
      }

      const redirectState = location.state as RedirectState | null;
      const redirectPath = resolveRedirectPath(redirectState?.from);

      void navigate(redirectPath, { replace: true });
    } catch {
      // Ошибки входа уже отражены в auth-store.
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

        <PasswordInput
          label="Пароль"
          value={password}
          onChange={(event) => setPassword(event.currentTarget.value)}
          autoComplete="current-password"
          required
          disabled={isDisabled}
        />

        {errorMessage ? (
          <Text size="sm" c="alarmCrit">
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
