import "@mantine/core/styles.css";

import { RouteProvider } from "./providers/RouterProvider/RouteProvider";
import { ThemeProvider } from "./providers/ThemeProvider";

function App() {
  return (
    <ThemeProvider>
      <RouteProvider />
    </ThemeProvider>
  );
}

export default App;
