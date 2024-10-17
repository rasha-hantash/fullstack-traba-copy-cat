import Image from "next/image";
import Sidebar from "@/components/Sidebar";
import MainContainer from "@/components/MainContainer";


export default function Home() {
  return (
    <div className="flex">
          <Sidebar />
          <MainContainer/>
    </div>
  );
}
