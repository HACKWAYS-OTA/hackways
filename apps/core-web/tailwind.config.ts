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
          50: "#faf8fd",
          100: "#eee7f6",
          500: "#887098",
          600: "#72548e",
          700: "#543d69",
        },
      },
    },
  },
  plugins: [],
};

export default config;
