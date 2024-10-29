import {
  Receipt,
  RefreshCcw,
} from "lucide-react";

interface Invoice {
  id: string;
  invoice_amount: number;
  start_date: string;
  end_date: string;
  status: string;
  invoice_name: string;
}

interface InvoicesProps {
  invoices: Invoice[] | null;
  isLoading: boolean;
}

export default function Invoices({ invoices, isLoading }: InvoicesProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  };

  if (isLoading) {
    return <div className="p-4">Loading...</div>;
  }

  if (!invoices) {
    return <div className="p-4">No invoices found</div>;
  }

  return (
    <div>
      <div className="dark:bg-black bg-gray-100 align-middle">
        <table className="min-w-full">
          <thead>
            <tr>
              <th scope="col" className="pl-4 dark:text-white py-1.5 pr-3 text-left text-sm font-semibold text-gray-500">
                Invoice ID
              </th>
              <th scope="col" className="dark:text-white py-1.5 px-3 text-left text-sm font-semibold text-gray-500">
                Start Date
              </th>
              <th scope="col" className="dark:text-white py-1.5 px-3 text-left text-sm font-semibold text-gray-500">
                End Date
              </th>
              <th scope="col" className="flex items-center dark:text-white py-1.5 px-3 text-left text-sm font-semibold text-gray-500">
                Invoice Description
              </th>
              <th scope="col" className="dark:text-white py-1.5 px-3 text-left text-sm font-semibold text-gray-500">
                Invoice Amount
              </th>
              <th scope="col" className="dark:text-white py-1.5 px-3 text-left text-sm font-semibold text-gray-500">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="dark:bg-black divide-y divide-gray-200 bg-white">
            {invoices.map((invoice) => (
              <tr key={invoice.id}>
                <td className="pl-4 dark:text-white whitespace-nowrap py-1.5 pr-3 text-sm text-gray-900">
                  {invoice.id}
                </td>
                <td className="dark:text-white whitespace-nowrap px-3 py-1.5 text-sm text-gray-900">
                  {formatDate(invoice.start_date)}
                </td>
                <td className="dark:text-white whitespace-nowrap px-3 py-1.5 text-sm text-gray-900">
                  {formatDate(invoice.end_date)}
                </td>
                <td className="dark:text-white whitespace-nowrap px-3 py-1.5 text-sm text-gray-900">
                  {invoice.invoice_name}
                </td>
                <td className="dark:text-white whitespace-nowrap px-3 py-1.5 text-sm text-gray-900">
                  {formatCurrency(invoice.invoice_amount)}
                </td>
                <td className="dark:text-white whitespace-nowrap px-3 py-1.5 text-sm text-gray-900">
                  <span
                    className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${
                      invoice.status.toLowerCase() === "unpaid"
                        ? "bg-red-50 text-red-700 ring-red-600/20"
                        : "bg-green-50 text-green-700 ring-green-600/20"
                    }`}
                  >
                    {invoice.status}
                  </span>
                </td>
                <td className="flex items-center justify-end pr-4 relative whitespace-nowrap py-1.5 text-right text-sm font-medium">
                  <a className={invoice.status.toLowerCase() === "paid" ? "pointer-events-none" : ""}>
                    <RefreshCcw
                      className={`mr-2 py-1 ${
                        invoice.status.toLowerCase() === "paid"
                          ? "text-gray-300 cursor-default"
                          : "text-gray-600 cursor-pointer hover:text-gray-900"
                      }`}
                    />
                  </a>
                  <a
                    href={invoice.status.toLowerCase() === "paid" ? "#" : "#actual-pay-link"}
                    className={`flex items-center px-1 rounded-md border ${
                      invoice.status.toLowerCase() === "paid"
                        ? "bg-gray-100 text-gray-400 cursor-default"
                        : "hover:bg-gray-100 text-gray-700"
                    }`}
                    onClick={(e) => invoice.status.toLowerCase() === "paid" && e.preventDefault()}
                  >
                    <Receipt
                      className={`py-1 ${
                        invoice.status.toLowerCase() === "paid"
                          ? "text-gray-400"
                          : "text-gray-600"
                      }`}
                    />
                    Pay
                  </a>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className="min-w-full border-b border-gray-200"></div>
      </div>
    </div>
  );
}