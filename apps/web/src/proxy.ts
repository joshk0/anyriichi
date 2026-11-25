import { NextResponse, type NextRequest } from "next/server";
import { locales, negotiateLocale } from "@/lib/i18n";

/** Redirect locale-less paths to /en or /ja based on Accept-Language. */
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasLocale = locales.some((l) => pathname === `/${l}` || pathname.startsWith(`/${l}/`));
  if (hasLocale) return;
  request.nextUrl.pathname = `/${negotiateLocale(request.headers.get("accept-language"))}${pathname}`;
  return NextResponse.redirect(request.nextUrl);
}

export const config = {
  matcher: ["/((?!_next|api|favicon.ico|.*\\..*).*)"],
};
