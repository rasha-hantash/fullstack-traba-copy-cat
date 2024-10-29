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

interface Invoice {
  id: string;
  invoice_amount: number;
  start_date: string;
  end_date: string;
  status: string;
  invoice_name: string;
}

export default function MainContainer() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [debounceTimeout, setDebounceTimeout] = useState<NodeJS.Timeout | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [invoiceData, setInvoiceData] = useState<Invoice[] | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Handle search and API call
  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    setError(null);
    
    // Clear existing timeout
    if (debounceTimeout) {
      clearTimeout(debounceTimeout);
    }

    // Set new timeout to debounce API calls
    const timeout = setTimeout(async () => {
      setIsLoading(true);
      try {
        const response = await fetch(`/api/invoices?search=${encodeURIComponent(query)}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error('Network response was not ok');
        }

        const data = await response.json();
        setInvoiceData(data);
        
      } catch (error) {
        console.error('Error fetching search results:', error);
        setError(error instanceof Error ? error.message : 'An error occurred');
      } finally {
        setIsLoading(false);
      }
    }, 300); // 300ms debounce delay

    setDebounceTimeout(timeout);
  };

  // Initial data fetch
  useEffect(() => {
    const fetchInitialData = async () => {
      setIsLoading(true);
      try {
        const response = await fetch('/api/invoices');
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        const data = await response.json();
        setInvoiceData(data);
      } catch (error) {
        console.error('Error fetching initial data:', error);
        setError(error instanceof Error ? error.message : 'An error occurred');
      } finally {
        setIsLoading(false);
      }
    };

    fetchInitialData();
  }, []);

  const toggleSidebar = () => {
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
    <div className="flex-auto border lg:rounded-md lg:m-2">
      <div
        className={`lg:hidden fixed inset-0 bg-black transition-opacity duration-300 ease-in-out ${
          isSidebarOpen
            ? "opacity-5 pointer-events-auto"
            : "opacity-0 pointer-events-none"
        } z-40`}
        onClick={toggleSidebar}
      />

      <div
        className={`lg:hidden z-50 fixed top-0 left-0 h-full w-64 bg-white transform transition-transform duration-300 ease-in-out ${
          isSidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <Sidebar />
      </div>

      <div className="px-4 flex border-b items-center">
        <div className="w-full">
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
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
              placeholder="Search your invoices..."
              className="h-full w-full border-0 py-0 pl-2 pr-0 text-gray-900 placeholder:text-gray-600 focus:ring-0 text-sm"
            />
          </form>
        </div>
        <div className="border-l flex items-center">
          <ThemeSwitcher />
          <Bell className="cursor-pointer py-1" />
        </div>
      </div>

      <div className="flex items-center justify-between px-4 py-1 border-b">
        <div>
          <h2 className="text-sm text-gray-900">All invoices</h2>
        </div>
        <div className="flex">
          <button className="hover:text-black text-gray-600 flex items-center text-xs">
            <ListFilter className="py-1" />
            Filter
          </button>
          <button className="hover:text-black text-gray-600 mx-1 flex items-center text-xs">
            <FileDown className="py-1" />
            Export
          </button>
          <button className="ml-1 flex shrink-0 items-center text-xs shadow-sm border hover:bg-gray-100 rounded-md">
            <SquarePen className="py-1" />
            <span className="pr-1">Create invoice</span>
          </button>
        </div>
      </div>

      {error ? (
        <div className="text-red-500 p-4">Error: {error}</div>
      ) : (
        <Invoices 
          invoices={invoiceData} 
          isLoading={isLoading} 
        />
      )}
    </div>
  );
}