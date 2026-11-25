import type { Metadata } from "next";
import { Inter, Space_Grotesk, Noto_Sans_JP } from "next/font/google";
import { locales } from "@/lib/i18n";
import { getDictionary } from "./dictionaries";
import { SiteHeader } from "@/components/site-header";
import "../globals.css";

const inter = Inter({ variable: "--font-inter", subsets: ["latin"] });
const grotesk = Space_Grotesk({ variable: "--font-grotesk", subsets: ["latin"] });
const notoJp = Noto_Sans_JP({ variable: "--font-noto-jp", weight: ["400", "500", "700"], preload: false });

export async function generateStaticParams() {
  return locales.map((lang) => ({ lang }));
}

export async function generateMetadata(): Promise<Metadata> {
  const dict = await getDictionary();
  return { title: dict.meta.title, description: dict.meta.description };
}

export default async function RootLayout({ children, params }: LayoutProps<"/[lang]">) {
  const { lang } = await params;
  return (
    <html lang={lang} className={`${inter.variable} ${grotesk.variable} ${notoJp.variable} h-full antialiased`}>
      <body className="min-h-full bg-ground font-sans text-ink">
        <SiteHeader lang={lang} />
        {children}
      </body>
    </html>
  );
}
