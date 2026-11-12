import { matchesCatalogSearch } from "@helpers/catalog-search";
import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { type LogAuditRule, logAuditRulesResponseSchema } from "@schemas/log-audit";
import {
  createContext,
  type PropsWithChildren,
  type ReactElement,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

type AuditValue = {
  rules: LogAuditRule[];
  visibleRules: LogAuditRule[];
  loading: boolean;
  error: boolean;
  search: string;
  setSearch: (value: string) => void;
  refresh: () => Promise<void>;
};

const AuditContext = createContext<AuditValue | undefined>(undefined);

/** Owns the validated audit catalog and its local search state. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider owns one catalog request and its search state.
export const AutomatedAuditProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { get } = useApi();
  const { requireSecurity } = useRbac();
  const [rules, setRules] = useState<LogAuditRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [search, setSearch] = useState("");
  const allowed = requireSecurity();

  const refresh = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Guard each async state update after navigation.
    async (isCurrent: () => boolean = () => true): Promise<void> => {
      if (!allowed) return;
      setLoading(true);
      try {
        const response = await get("/api/security/log-audit/rules").execute();
        if (!response.ok) throw new Error("log audit request failed");
        const catalog = logAuditRulesResponseSchema.parse(await response.json());
        if (isCurrent()) {
          setRules(catalog.data.rules);
          setError(false);
        }
      } catch {
        if (isCurrent()) setError(true);
      } finally {
        if (isCurrent()) setLoading(false);
      }
    },
    [allowed, get],
  );

  useEffect(() => {
    let current = true;
    void refresh(() => current);
    return () => {
      current = false;
    };
  }, [refresh]);
  const visibleRules = rules.filter((rule) =>
    matchesCatalogSearch(search, [
      rule.id,
      rule.name,
      rule.explanation,
      rule.source,
      rule.category,
      rule.severity,
      rule.instance,
      rule.threshold,
      rule.window,
      ...rule.possible_reactions,
    ]),
  );
  return (
    <AuditContext.Provider
      value={{ rules, visibleRules, loading, error, search, setSearch, refresh }}
    >
      {children}
    </AuditContext.Provider>
  );
};

/** Reads audit catalog data and search controls from its page provider. */
export const useAutomatedAudit = (): AuditValue => {
  const value = useContext(AuditContext);
  if (!value) throw new Error("useAutomatedAudit must be used inside AutomatedAuditProvider");
  return value;
};
