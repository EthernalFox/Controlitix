import { IconAlertTriangle } from "@tabler/icons-react";

import { Button } from "@shared/ui";

import styles from "./SaveErrorBanner.module.css";

interface SaveErrorBannerProps {
  onRetry: () => void;
}

export const SaveErrorBanner = ({ onRetry }: SaveErrorBannerProps) => {
  return (
    <div className={styles.saveErrorBanner}>
      <IconAlertTriangle size={14} />
      <span>Не удалось сохранить</span>
      <Button size="xs" variant="ghost" onClick={onRetry}>
        Повторить
      </Button>
    </div>
  );
};
