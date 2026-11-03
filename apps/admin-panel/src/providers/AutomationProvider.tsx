import { matchesCatalogSearch } from "@helpers/catalog-search";
import { useAutomation } from "@providers/useAutomation";
import {
  createContext,
  type PropsWithChildren,
  type ReactElement,
  useContext,
  useState,
} from "react";

type AutomationValue = ReturnType<typeof useAutomation> & {
  search: string;
  setSearch: (value: string) => void;
  visibleRules: ReturnType<typeof useAutomation>["rules"];
  visibleActive: ReturnType<typeof useAutomation>["active"];
};

const AutomationContext = createContext<AutomationValue | undefined>(undefined);

/** Owns the automation snapshot and a search shared by rules and active effects. */
export const AutomationProvider = ({ children }: PropsWithChildren): ReactElement => {
  const automation = useAutomation();
  const [search, setSearch] = useState("");
  const visibleRules = automation.rules.filter((rule) =>
    matchesCatalogSearch(search, [
      rule.id,
      rule.name,
      rule.description,
      rule.explanation,
      rule.action_kind,
      rule.action_name,
      ...(rule.required_operands ?? []),
      rule.last_outcome,
    ]),
  );
  const visibleActive = automation.active.filter((incident) =>
    matchesCatalogSearch(search, [
      incident.incident_id,
      incident.rule.id,
      incident.rule.name,
      incident.rule.description,
      incident.rule.explanation,
      incident.target,
      incident.mitigation,
      incident.last_outcome,
    ]),
  );
  return (
    <AutomationContext.Provider
      value={{ ...automation, search, setSearch, visibleRules, visibleActive }}
    >
      {children}
    </AutomationContext.Provider>
  );
};

/** Reads automation data and the page search from their provider. */
export const useAutomationContext = (): AutomationValue => {
  const value = useContext(AutomationContext);
  if (!value) throw new Error("useAutomationContext must be used inside AutomationProvider");
  return value;
};
