import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#FAF7FE",
          100: "#F1EAFA",
          200: "#E4D8F3",
          300: "#CBBADF",
          400: "#9A7CBB",
          500: "#1F1F23",
          600: "#343139",
          700: "#4E3B63",
          800: "#3B2C4D",
          900: "#211A2A",
        },
        surface: {
          canvas: "#FFFFFF",
          pure: "#FFFFFF",
          subdued: "#FFFFFF",
        },
        ink: {
          heavy: "#1F1F1F",
          body: "#4B4650",
          muted: "#706977",
        },
        border: {
          hairline: "#DFD7E8",
          emphasized: "#CBBADC",
        },
      },
      fontFamily: {
        sans: ["Google Sans Flex"],
        heading: ["FG Futurist"],
        mono: ["var(--font-geist-mono)", "Geist Mono", "ui-monospace", "SFMono-Regular", "monospace"],
      },
      borderRadius: {
        pill: "9999px",
        xl: "12px",
        "2xl": "16px",
        "3xl": "24px",
      },
      boxShadow: {
        pill: "0 1px 3px rgba(14, 18, 30, 0.08)",
        "pill-hover": "0 4px 12px rgba(31, 31, 35, 0.12)",
        "pill-primary": "0 2px 8px rgba(31, 31, 35, 0.18)",
        modal: "0 20px 40px -15px rgba(14, 18, 30, 0.15)",
      },
    },
  },
  plugins: [],
};

export default config;
