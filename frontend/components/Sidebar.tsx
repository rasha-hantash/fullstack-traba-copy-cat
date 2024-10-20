"use client";
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

// todo: look up and see if i actually need the classNames function
function classNames(...classes: string[]) {
    return classes.filter(Boolean).join(" ");
}

export default function Sidebar() {
    return (
        <>
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
        </>
    )



}