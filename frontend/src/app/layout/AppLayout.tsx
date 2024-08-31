import { Header } from "@shared/ui/Header";
import { Navbar } from "@widgets/Navbar";
import { Outlet } from "react-router-dom";
import styles from "./AppLayout.module.css";

const AppLayout = () => {
  return (
    <div className={styles.layout}>
      <Header title={"Заголовок"} />
      <main className={styles.main}>
        <Outlet />
        <Navbar />
      </main>
    </div>
  );
};

export default AppLayout;
