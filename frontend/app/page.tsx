import Image from "next/image";
import Sidebar from "@/components/Sidebar";
import MainContainer from "@/components/MainContainer";

export default function Home() {
  return (
    <div className="lg:flex">
      <div className="hidden lg:block w-60 h-screen dark:bg-black">
        <Sidebar />
      </div>
      <MainContainer />
    </div>
  );
}
