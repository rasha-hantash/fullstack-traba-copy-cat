"use client";
import {
  Search,
  ListFilter,
  FileDown,
  SquarePen,
  Bell,
  PanelLeft,
} from "lucide-react";
import ThemeSwitcher from "./ThemeSwitcher";
function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(" ");
}

export default function MainContainer() {
  return (
    <div className="flex-auto border lg:rounded-md lg:m-2 ">
      {/* todo: create a search bar component */}
      {/* do i neet method.GET ?  */}
      <div className="flex border-b items-center">
        <div className="ml-2 flex-auto w-full">
          <form
            action="#"
            method="GET"
            className="h-10 relative flex flex-1 items-center"
          >
            <PanelLeft className="lg:hidden ml-2 h-4 w-4 text-gray-600" />

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
