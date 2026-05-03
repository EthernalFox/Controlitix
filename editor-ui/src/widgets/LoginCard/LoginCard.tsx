import { useState } from "react";

import { Button, PasswordInput, Stack, TextInput } from "@shared/ui";
import { LogoMark } from "@widgets/LogoMark";

import styles from "./LoginCard.module.css";

export const LoginCard = () => {
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    await new Promise((resolve) => {
      window.setTimeout(resolve, 500);
    });

    if (!login.trim() || !password.trim()) {
      setError("Укажите логин и пароль");
      setIsSubmitting(false);
      return;
    }

    setIsSubmitting(false);
  };

  return (
    <form className={styles.card} onSubmit={onSubmit}>
      <div className={styles.header}>
        <LogoMark />
        <h1 className={styles.title}>С возвращением, инженер</h1>
      </div>

      <Stack gap="sm">
        <TextInput
          label="Логин"
          value={login}
          onChange={(event) => setLogin(event.currentTarget.value)}
          autoFocus
        />
        <PasswordInput
          label="Пароль"
          value={password}
          onChange={(event) => setPassword(event.currentTarget.value)}
        />
        {error && <p className={styles.error}>{error}</p>}
        <Button type="submit" loading={isSubmitting} fullWidth>
          Войти
        </Button>
      </Stack>
    </form>
  );
};
