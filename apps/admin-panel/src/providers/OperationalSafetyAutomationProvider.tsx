import { matchesCatalogSearch } from "@helpers/catalog-search";
import { useRbac } from "@hooks/useRbac";
import { useOperationalSafetyAutomation } from "@providers/useOperationalSafetyAutomation";
import {
  createContext,
  type PropsWithChildren,
  type ReactElement,
  useContext,
  useState,
} from "react";

const OperationalSafetyAutomationContext = createContext<
  | (ReturnType<typeof useOperationalSafetyAutomation> & {
      search: string;
      setSearch: (value: string) => void;
    })
  | undefined
>(undefined);

/** Owns the operational safety snapshot and search shared by rules and tasks. */
export const OperationalSafetyAutomationProvider = ({
  children,
}: PropsWithChildren): ReactElement => {
  const { requireAdmin } = useRbac();
  const automation = useOperationalSafetyAutomation(requireAdmin());
  const [search, setSearch] = useState("");
  const rules = automation.rules.filter((rule) =>
    matchesCatalogSearch(search, [rule.id, rule.name, rule.output_id_prefix]),
  );
  const tasks = automation.tasks.filter((task) =>
    matchesCatalogSearch(search, [
      task.id,
      task.rule_id,
      task.correlation_key,
      task.level,
      task.last_outcome,
      ...task.active_alert_ids,
    ]),
  );
  return (
    <OperationalSafetyAutomationContext.Provider
      value={{ ...automation, rules, tasks, search, setSearch }}
    >
      {children}
    </OperationalSafetyAutomationContext.Provider>
  );
};

/** Reads operational safety automation data and page search from its provider. */
export const useOperationalSafetyAutomationContext = () => {
  const value = useContext(OperationalSafetyAutomationContext);
  if (!value)
    throw new Error("useOperationalSafetyAutomationContext must be used inside its provider");
  return value;
};
