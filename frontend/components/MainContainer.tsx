"use client";
import { useState } from "react";
import {
  Search,
  ListFilter,
  FileDown,
  SquarePen,
  Bell,
  PanelLeft,
} from "lucide-react";
import ThemeSwitcher from "./ThemeSwitcher";
import Sidebar from "./Sidebar";
function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(" ");
}

export default function MainContainer() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  const toggleSidebar = () => {
    console.log(isSidebarOpen);
    setIsSidebarOpen(!isSidebarOpen);
  };

  return (
    <div className="h-screen flex-auto border lg:rounded-md lg:m-2 ">
      {/* todo: create a search bar component */}
      {/* do i neet method.GET ?  */}
            {/* Sidebar */}
            <div
        className={`z-50 fixed top-0 right-0 h-full w-64 bg-gray-200 p-4 transform transition-transform duration-300 ease-in-out ${
          isSidebarOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <h2 className="text-xl font-semibold mb-4">Sidebar</h2>
        <p>This is the sidebar content.</p>
        <button
          onClick={toggleSidebar}
          className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition-colors"
        >
          Close Sidebar
        </button>
      </div>

      <div className="flex border-b items-center">
        <div className="ml-2 flex-auto w-full">
          <form
            action="#"
            method="GET"
            className="h-10 relative flex flex-1 items-center"
          >
            <PanelLeft  onClick={toggleSidebar} className="lg:hidden ml-2 h-4 w-4 text-gray-600" />

            <label htmlFor="search-field" className="sr-only">
              Search
            </label>
            <Search aria-hidden="true" className="pl-2 h-6 w-6 text-gray-600" />

            <input
              id="search-field"
              name="search"
              // type="search"
              placeholder="Search your invoices..."
              className="h-full w-full border-0 py-0 pl-2 pr-0 text-gray-900 placeholder:text-gray-600 focus:ring-0 sm:text-sm"
            />
          </form>
        </div>
        <div className="border-l flex items-center mr-2">
          <ThemeSwitcher />
          <button className="flex items-center text-xs">
            <Bell className="py-1" />
          </button>
        </div>
      </div>

      {/* Subheader */}
      {/* todo: include gap in spacing inbetween icons: example gap-x-2*/}
      <div className="flex items-center justify-between px-1 m:px-2 lg:px-4 py-1 border-b">
        <div>
          <h2 className="text-sm  text-gray-900">All invoices</h2>
        </div>
        <div className="flex">
          <button className="hover:text-black text-gray-700 flex items-center text-xs">
            <ListFilter className="py-1" />
            Filter
          </button>
          <button className="hover:text-black text-gray-700 mx-1 flex items-center text-xs">
            <FileDown className="py-1" />
            Export
          </button>
          <button className="ml-1 flex shrink-0 items-center text-xs shadow-sm border hover:bg-gray-100  rounded-md">
            <SquarePen className="py-1" />
            <span className="pr-1">Create invoice</span>
          </button>
        </div>
      </div>
    </div>
  );
}
