import faStaticMessages from './fa-static.json';
import faScriptMessages from './fa-script.json';

const enStaticMessages = Object.fromEntries(
  [...Object.keys(faStaticMessages), ...Object.keys(faScriptMessages)].map((sourceText) => [sourceText, sourceText]),
);

export const messages = {
  en: {
    language: {
      label: 'Language',
      english: 'English',
      persian: 'فارسی',
    },
    brand: {
      tagline: 'Print shop control',
      localWorkspace: 'Local workspace',
    },
    navigation: {
      primary: 'Primary navigation',
      close: 'Close navigation',
      sections: {
        workspace: 'Workspace',
        catalog: 'Catalog & purchasing',
        finance: 'Finance',
        insights: 'Insights & setup',
      },
      views: {
        Dashboard: 'Dashboard',
        Orders: 'Orders',
        Production: 'Production',
        Customers: 'Customers',
        Services: 'Services',
        Materials: 'Materials',
        Machines: 'Machines',
        Purchases: 'Purchases',
        Suppliers: 'Suppliers',
        Accounting: 'Accounting',
        Invoices: 'Invoices',
        Checks: 'Checks',
        Loans: 'Loans',
        Owners: 'Owners',
        Reports: 'Reports',
        Settings: 'Settings',
      },
      unavailable: 'Workspace view unavailable',
      unavailableDescription: '{view} is not available yet. Return to the dashboard to continue.',
      backToDashboard: 'Back to dashboard',
    },
    workspace: {
      eyebrow: 'Atropaten workspace',
      description: 'This workspace is ready for the next product milestone.',
    },
    toolbar: {
      expandSidebar: 'Expand sidebar',
      collapseSidebar: 'Collapse sidebar',
      search: 'Search anything...',
      globalSearch: 'Global search',
      searching: 'Searching workspace…',
      noResults: 'No matching records.',
      currency: 'Display currency',
      toman: 'Toman',
      rial: 'Rial',
      quickActions: 'Quick actions',
      newOrder: 'New order',
    },
    notifications: {
      trigger: 'Notifications, {count} unread',
      markRead: 'Mark {title} as read',
    },
    ...enStaticMessages,
  },
  fa: {
    language: {
      label: 'زبان',
      english: 'انگلیسی',
      persian: 'فارسی',
    },
    brand: {
      tagline: 'مدیریت چاپخانه',
      localWorkspace: 'فضای کاری محلی',
    },
    navigation: {
      primary: 'ناوبری اصلی',
      close: 'بستن ناوبری',
      sections: {
        workspace: 'فضای کاری',
        catalog: 'کاتالوگ و خرید',
        finance: 'مالی',
        insights: 'گزارش‌ها و تنظیمات',
      },
      views: {
        Dashboard: 'داشبورد',
        Orders: 'سفارش‌ها',
        Production: 'تولید',
        Customers: 'مشتریان',
        Services: 'خدمات',
        Materials: 'مواد اولیه',
        Machines: 'دستگاه‌ها',
        Purchases: 'خریدها',
        Suppliers: 'تأمین‌کنندگان',
        Accounting: 'حسابداری',
        Invoices: 'فاکتورها',
        Checks: 'چک‌ها',
        Loans: 'وام‌ها',
        Owners: 'شرکا',
        Reports: 'گزارش‌ها',
        Settings: 'تنظیمات',
      },
      unavailable: 'صفحهٔ کاری در دسترس نیست',
      unavailableDescription: 'صفحهٔ «{view}» هنوز در دسترس نیست. برای ادامه به داشبورد برگردید.',
      backToDashboard: 'بازگشت به داشبورد',
    },
    workspace: {
      eyebrow: 'فضای کاری آتروپاتن',
      description: 'فضای کاری برای مرحلهٔ بعدی محصول آماده است.',
    },
    toolbar: {
      expandSidebar: 'باز کردن نوار کناری',
      collapseSidebar: 'بستن نوار کناری',
      search: 'جستجو در همه‌جا...',
      globalSearch: 'جستجوی سراسری',
      searching: 'در حال جستجو در فضای کاری…',
      noResults: 'موردی پیدا نشد.',
      currency: 'واحد نمایش ارز',
      toman: 'تومان',
      rial: 'ریال',
      quickActions: 'دسترسی سریع',
      newOrder: 'سفارش جدید',
    },
    notifications: {
      trigger: 'اعلان‌ها، {count} خوانده‌نشده',
      markRead: 'علامت‌گذاری {title} به‌عنوان خوانده‌شده',
    },
    ...faScriptMessages,
    ...faStaticMessages,
  },
};

export type Locale = keyof typeof messages;
