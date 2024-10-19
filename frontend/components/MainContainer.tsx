"use client";
import { useState, useEffect } from "react";
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
import Invoices from "./Invoices";
function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(" ");
}

import {
  CalendarDays,
  Users,
  LayoutGrid,
  ReceiptText,
  FileSpreadsheet,
  UserCog,
  BriefcaseBusiness,
} from "lucide-react";

const navigation = [
  { name: "Dashboard", href: "#", icon: LayoutGrid, current: true },
  { name: "Calendar", href: "#", icon: CalendarDays, current: false },
  { name: "Workers", href: "#", icon: Users, current: false },
  { name: "Timesheet", href: "#", icon: FileSpreadsheet, current: false },
  {
    name: "Invoices",
    href: "#",
    icon: ReceiptText,
    submenu: [
      { name: "Draft", href: "#draft" },
      { name: "Outstanding", href: "#outstanding" },
      { name: "Past due", href: "#past-due" },
      { name: "Paid", href: "#paid" },
    ],
    current: false,
  },
  { name: "Post shift", href: "#", icon: BriefcaseBusiness, current: false },
];


export default function MainContainer() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  const toggleSidebar = () => {
    console.log(isSidebarOpen);
    setIsSidebarOpen(!isSidebarOpen);
  };

  useEffect(() => {
    const mediaQuery = window.matchMedia('(min-width: 1024px)');

    const handleMediaQueryChange = (e: MediaQueryListEvent) => {
      if (e.matches) {
        setIsSidebarOpen(false);
      }
    };

    mediaQuery.addEventListener('change', handleMediaQueryChange);

    return () => {
      mediaQuery.removeEventListener('change', handleMediaQueryChange);
    };
  }, []);

  return (
    <div className="flex-auto border lg:rounded-md lg:m-2 ">
      {/* todo: create a search bar component */}
      {/* do i neet method.GET ?  */}

      <div
        className={`lg:hidden fixed inset-0 bg-black  transition-opacity duration-300 ease-in-out ${
          isSidebarOpen
            ? "opacity-20 pointer-events-auto"
            : "opacity-0 pointer-events-none"
        } z-40`}
        onClick={toggleSidebar}
      />

      {/* Sidebar */}
      <div
        className={`lg:hidden z-50 fixed top-0 left-0 h-full w-64 bg-gray-50  transform transition-transform duration-300 ease-in-out ${
          isSidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        {/* todo: figure out how to just use the sidebar component */}
            <nav>
            <div className="px-4 h-8 pt-1">
              <img
                alt="Your Company"
                src="https://tailwindui.com/plus/img/logos/mark.svg?color=indigo&shade=600"
                className="h-8 w-auto"
              />
            </div>
            <ul role="list" className="mt-2 px-4">
                <li>
                  <ul role="list">
                    {navigation.map((item) => (
                      <li key={item.name}>
                        <a
                          href={item.href}
                          className={classNames(
                            item.current
                              ? "text-black bg-gray-100"
                              : "text-gray-600 hover:bg-gray-50 hover:text-black",
                            "group flex gap-x-1 rounded-md text-sm leading-6"
                          )}
                        >
                          <item.icon
                            aria-hidden="true"
                            className={classNames(
                              item.current
                                ? "text-gray-600 group-hover:text-black"
                                : "text-gray-600 group-hover:text-black",
                              "h-6 w-6 shrink-0 p-1"
                            )}
                          />
                          <span className="group-hover:text-black">
                            {item.name}
                          </span>
                        </a>
                        {item.submenu && (
                          <ul className="my-1 ml-3 ">
                            {item.submenu.map((subitem) => (
                              <li key={subitem.name} className="border-l">
                                <a
                                  href={subitem.href}
                                  className="font-thin text-xs hover:rounded-md hover:bg-gray-50  ml-1 pl-3 py-1 border-gray-200 text-black block"
                                >
                                  {subitem.name}
                                </a>
                              </li>
                            ))}
                          </ul>
                        )}
                      </li>
                    ))}
                  </ul>
                </li>
                <li className="border-t mt-2 py-1">
                  <a
                    href="#"
                    className="text-gray-600 hover:bg-gray-50 hover:text-black group flex gap-x-3 rounded-md text-sm leading-6"
                  >
                    <UserCog
                      aria-hidden="true"
                      className="text-gray-600 group-hover:text-black h-6 w-6 shrink-0 p-1"
                    />
                    <span className="group-hover:text-black">
                            Settings
                    </span>
                  </a>
                </li>
              </ul>
            </nav>
      </div>

      <div className="px-4 flex border-b items-center">
        <div className="flex-auto w-full">
          <form
            action="#"
            method="GET"
            className="h-10 relative flex flex-1 items-center"
          >
            <PanelLeft
              onClick={toggleSidebar}
              className="lg:hidden mr-1 h-4 w-4 text-gray-600"
            />

            <label htmlFor="search-field" className="sr-only">
              Search
            </label>
            <Search aria-hidden="true" className="h-4 w-4 text-gray-600" />

            <input
              id="search-field"
              name="search"
              // type="search"
              placeholder="Search your invoices..."
              className="h-full w-full border-0 py-0 pl-2 pr-0 text-gray-900 placeholder:text-gray-600 focus:ring-0 sm:text-sm"
            />
          </form>
        </div>
        <div className="border-l flex items-center">
          <ThemeSwitcher />
          <Bell className="cursor-pointer py-1" />
        </div>
      </div>

      {/* Subheader */}
      <div className="flex items-center justify-between px-4 py-1 border-b">
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
      <Invoices />
    </div>
  );
}
