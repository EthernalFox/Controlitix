import "@mantine/core/styles.css";

import { AuthProvider } from "./providers/AuthProvider";
import { RouteProvider } from "./providers/RouterProvider/RouteProvider";
import { ThemeProvider } from "./providers/ThemeProvider";

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <RouteProvider />
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
