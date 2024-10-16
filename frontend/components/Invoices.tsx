import { ListFilter, ArrowUpRight, SquarePen, Receipt, RefreshCcw } from "lucide-react";

const people = [
  {
    name: "Lindsay Walton",
    title: "Front-end Developer",
    email: "lindsay.walton@example.com",
    role: "Member",
  },
  // More people...
];

export default function Example() {
  return (
    <div>
      <div className="mb-1 sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-sm  text-gray-900">All invoices</h1>
        </div>
       
        <button className="flex items-center text-xs ml-1"> <ListFilter className="py-1"/> Filter</button>
        <button className="flex items-center text-xs  ml-4">
            <ArrowUpRight className="py-1"/>
          Export
        </button>
        <button className="shadow-sm border flex hover:bg-gray-100  rounded-md items-center pr-1 text-xs  ml-4">
            <SquarePen className="py-1"/>
          Create invoice
        </button>
      </div>

      <div className="flow-root">
        <div className="-mx-4  sm:-mx-6 lg:-mx-8">
          <div className="border-t border-gray-200"></div>

          <div className="inline-block min-w-full align-middle">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="mx-2 min-w-full divide-y divide-gray-100">
                <tr>
                  <th
                    scope="col"
                    className="py-1.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6 lg:pl-8"
                  >
                    Invoice ID
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 px-3 text-left text-sm font-semibold text-gray-900"
                  >
                    Start Date
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 px-3 text-left text-sm font-semibold text-gray-900"
                  >
                    End Date
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 px-3  text-left text-sm font-semibold text-gray-900"
                  >
                    Total Shifts
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 px-3  text-left text-sm font-semibold text-gray-900"
                  >
                    Invoice Amount
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 px-3 text-left text-sm font-semibold text-gray-900"
                  >
                    Status
                  </th>
                  <th
                    scope="col"
                    className="py-1.5 relative pl-3 pr-4 sm:pr-6 lg:pr-8"
                  >
                    <span className="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {people.map((person) => (
                  <tr key={person.email}>
                    <td className="whitespace-nowrap py-1.5 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8">
                      {person.name}
                    </td>
                    <td className="whitespace-nowrap px-3 py-1.5 text-sm text-gray-500">
                      {person.title}
                    </td>
                    <td className="whitespace-nowrap px-3 py-1.5 text-sm text-gray-500">
                      {person.email}
                    </td>
                    <td className="whitespace-nowrap px-3 py-1.5 text-sm text-gray-500">
                      {person.role}
                    </td>
                    <td className="whitespace-nowrap px-3 py-1.5 text-sm text-gray-500">
                      $40
                    </td>
                    <td className="whitespace-nowrap px-3 py-1.5 text-sm text-gray-500">
                      <span className="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20">
                        Active
                      </span>
                    </td>
                    <td className="flex items-center relative whitespace-nowrap py-1.5 pr-4 text-right text-sm font-medium sm:pr-6 lg:pr-8">
                      <a><RefreshCcw className="cursor-pointer mr-2 py-1"/></a>
                      <a
                        href="#"
                        className="flex items-center px-1 rounded-md border hover:bg-gray-100"
                      >
                        <Receipt className="py-1"/>
                        Pay
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
