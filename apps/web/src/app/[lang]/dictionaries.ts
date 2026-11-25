import { lang } from "next/root-params";
import { notFound } from "next/navigation";
import { hasLocale } from "@/lib/i18n";

const dictionaries = {
  en: () => import("./dictionaries/en.json").then((m) => m.default),
  ja: () => import("./dictionaries/ja.json").then((m) => m.default),
};

export type Dictionary = Awaited<ReturnType<(typeof dictionaries)["en"]>>;

export const getDictionary = async (): Promise<Dictionary> => {
  const locale = await lang();
  if (!hasLocale(locale)) notFound();
  return dictionaries[locale]();
};
