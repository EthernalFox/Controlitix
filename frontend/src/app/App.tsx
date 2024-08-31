import { StoreProvider } from "./Providers/StoreProvider";
import Providers from "./Providers/Providers";
import { RouteProvider } from "./Providers/RouteProvider";

function App() {
  return (
    <>
      <Providers providers={[RouteProvider, StoreProvider]} />
    </>
  );
}

export default App;
