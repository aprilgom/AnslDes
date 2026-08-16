import { createContext, type ReactNode, useContext } from "react";

import type { DesignSystemRuntime } from "@ansldes/core";

const DesignSystemContext = createContext<DesignSystemRuntime | null>(null);
const FontWeightModeContext = createContext<FontWeightMode>("numeric");

export type FontWeightMode = "normal" | "numeric";

export function DesignSystemProvider({
  children,
  fontWeightMode = "numeric",
  runtime,
}: {
  children: ReactNode;
  fontWeightMode?: FontWeightMode;
  runtime: DesignSystemRuntime;
}) {
  return (
    <DesignSystemContext.Provider value={runtime}>
      <FontWeightModeContext.Provider value={fontWeightMode}>
        {children}
      </FontWeightModeContext.Provider>
    </DesignSystemContext.Provider>
  );
}

export function useNativeFontWeightMode(): FontWeightMode {
  return useContext(FontWeightModeContext);
}

export function useDesignSystem(): DesignSystemRuntime {
  const runtime = useContext(DesignSystemContext);
  if (!runtime) throw new Error("DesignSystemProvider is required");
  return runtime;
}
