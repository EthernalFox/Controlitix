import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/notifications/styles.css";
import { useEffect } from "react";

import { AuthProvider } from "@/app/providers/AuthProvider";
import { RouterProvider } from "@/app/providers/RouterProvider";
import { ThemeProvider } from "@/app/providers/ThemeProvider";
import { useAuthStore } from "@/features/auth";
import { realtimeClient } from "@/shared/modules/realtime";

function App() {
  const authStatus = useAuthStore((state) => state.status);
  const accessToken = useAuthStore((state) => state.accessToken);

  useEffect(() => {
    if (authStatus === "authenticated" && accessToken) {
      realtimeClient.start();
    } else {
      realtimeClient.stop();
    }
  }, [accessToken, authStatus]);

  useEffect(() => {
    return () => {
      realtimeClient.stop();
    };
  }, []);

  return (
    <ThemeProvider>
      <AuthProvider>
        <RouterProvider />
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
