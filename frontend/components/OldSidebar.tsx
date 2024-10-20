"use client";
import { useState } from "react";
import {
  Dialog,
  DialogBackdrop,
  DialogPanel,
  Menu,
  MenuButton,
  MenuItem,
  MenuItems,
  TransitionChild,
} from "@headlessui/react";

import {
  CalendarDays,
  Users,
  LayoutGrid,
  ReceiptText,
  FileSpreadsheet,
  UserCog,
  BriefcaseBusiness,
  PanelLeft,
  Bell,
} from "lucide-react";

import { XMarkIcon } from "@heroicons/react/24/outline";
import {
  ChevronDownIcon,
  MagnifyingGlassIcon,
} from "@heroicons/react/20/solid";
import ThemeSwitcher from "./ThemeSwitcher";
import Invoices from "./Invoices";

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


// todo: look up and see if i actually need the classNames function
function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(" ");
}

export default function OldSidebar() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <>
      {/*
        This example requires updating your template:

        ```
        <html class="h-full bg-white">
        <body class="h-full">
        ```
      */}
      <div className="inline-block min-w-full ">
        <Dialog
          open={sidebarOpen}
          onClose={setSidebarOpen}
          className="relative z-50 lg:hidden"
        >
          <DialogBackdrop
            transition
            className="fixed inset-0 bg-gray-900/80 transition-opacity duration-300 ease-linear data-[closed]:opacity-0"
          />

          <div className="fixed inset-0 flex">
            <DialogPanel
              transition
              className="relative mr-16 flex w-full max-w-xs flex-1 transform transition duration-300 ease-in-out data-[closed]:-translate-x-full"
            >
              <TransitionChild>
                <div className="absolute left-full top-0 flex w-16 justify-center pt-5 duration-300 ease-in-out data-[closed]:opacity-0">
                  <button
                    type="button"
                    onClick={() => setSidebarOpen(false)}
                    className="-m-2.5 p-2.5"
                  >
                    <span className="sr-only">Close sidebar</span>
                    <XMarkIcon
                      aria-hidden="true"
                      className="h-6 w-6 text-white"
                    />
                  </button>
                </div>
              </TransitionChild>
              {/* Sidebar component, swap this element with another sidebar if you like */}
              <div className="flex grow flex-col gap-y-1 overflow-y-auto bg-white px-6 pb-5">
                <div className="flex h-16 shrink-0 items-center">
                  <img
                    alt="Your Company"
                    src="https://tailwindui.com/plus/img/logos/mark.svg?color=indigo&shade=600"
                    className="h-8 w-auto"
                  />
                </div>
                <nav className="flex flex-1 flex-col">
                  <ul role="list" className="flex flex-1 flex-col gap-y-7">
                    <li>
                      <ul role="list" className=" -mx-2 space-y-1">
                        {navigation.map((item) => (
                          <li key={item.name}>
                            <a
                              href={item.href}
                              className={classNames(
                                item.current
                                  ? "text-black bg-gray-100"
                                  : "text-gray-600 hover:bg-gray-50 hover:text-black",
                                "group flex gap-x-3 rounded-md text-sm leading-6"
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
                    <li className="border-t">
                      <a
                        href="#"
                        className="group flex gap-x-3 rounded-md text-sm leading-6 text-gray-700 hover:bg-gray-50"
                      >
                        <UserCog
                          aria-hidden="true"
                          className="h-6 w-6 shrink-0 p-1 text-black "
                        />
                        Settings
                      </a>
                    </li>
                  </ul>
                </nav>
              </div>
            </DialogPanel>
          </div>
        </Dialog>

        {/* Static sidebar for desktop */}
        <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-52 lg:flex-col">
          {/* Sidebar component, swap this element with another sidebar if you like */}
          <div className="flex grow flex-col gap-y-5 overflow-y-auto border-r border-gray-200 bg-white px-6 pb-4">
            <div className="flex h-16 shrink-0 items-center">
              <img
                alt="Your Company"
                src="https://tailwindui.com/plus/img/logos/mark.svg?color=indigo&shade=600"
                className="h-8 w-auto"
              />
            </div>
            <nav className="flex flex-1 flex-col">
              <ul role="list" className="flex flex-1 flex-col gap-y-7">
                <li>
                  <ul role="list" className="-mx-2 space-y-1">
                    {navigation.map((item) => (
                      <li key={item.name}>
                        <a
                          href={item.href}
                          className={classNames(
                            item.current
                              ? "text-black bg-gray-100"
                              : "text-gray-600 hover:bg-gray-50 hover:text-black",
                            "group flex gap-x-3 rounded-md text-sm leading-6"
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
                <li className="border-t">
                  <a
                    href="#"
                    className="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm leading-6 text-gray-700 hover:bg-gray-50 "
                  >
                    <UserCog
                      aria-hidden="true"
                      className="h-6 w-6 shrink-0 text-sm text-black p-1"
                    />
                    Settings
                  </a>
                </li>
              </ul>
            </nav>
          </div>
        </div>

        <div className="lg:pl-52">
          <div className="sticky top-0 z-40 flex h-10 shrink-0 items-center gap-x-4 border-b border-gray-200 bg-white shadow-sm sm:gap-x-6 pl-6">
            <button
              type="button"
              onClick={() => setSidebarOpen(true)}
              className="-m-2.5  text-gray-700 lg:hidden"
            >
              <span className="sr-only">Open sidebar</span>
              <PanelLeft aria-hidden="true" className=" p-1" />
            </button>

            {/* Separator */}
            <div
              aria-hidden="true"
              className="h-6 w-px bg-gray-200 lg:hidden"
            />

            <div className="flex flex-1  self-stretch">
              <form action="#" method="GET" className="relative flex flex-1">
                <label htmlFor="search-field" className="sr-only">
                  Search
                </label>
                <MagnifyingGlassIcon
                  aria-hidden="true"
                  className="pointer-events-none absolute inset-y-0 left-0 h-full w-5 text-gray-400"
                />
                <input
                  id="search-field"
                  name="search"
                  type="search"
                  placeholder="Search..."
                  className="block h-full w-full border-0 py-0 pl-8 pr-0 text-gray-900 placeholder:text-gray-400 focus:ring-0 sm:text-sm"
                />
              </form>
              <div className="flex items-center p-2">
                {/* Separator */}
                <div
                  aria-hidden="true"
                  className="hidden lg:block lg:h-6 lg:w-px lg:bg-gray-200"
                />
                <button
                  type="button"
                  className="ml-2 text-gray-700 hover:text-gray-500"
                >
                  <span className="sr-only">View notifications</span>
                  <Bell aria-hidden="true" className=" h-4 w-4" />
                </button>
                <ThemeSwitcher />
                {/* Profile dropdown  removed */} 
              </div>
            </div>
          </div>

          <main className="py-1">
            <div className="px-4 sm:px-6 lg:px-8">
              <Invoices />
            </div>
          </main>
        </div>
      </div>
    </>
  );
}

// todo think about how to reduce code duplication
