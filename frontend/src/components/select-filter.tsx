import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type SelectFilterProps<Value extends string> = {
  label: string;
  value?: Value;
  options: readonly { value: Value; label: string }[];
  onChange: (value: Value | undefined) => void;
};

export function SelectFilter<Value extends string>({
  label,
  value,
  options,
  onChange,
}: SelectFilterProps<Value>) {
  return (
    <Select
      items={[{ value: "all", label }, ...options]}
      value={value ?? "all"}
      onValueChange={(selected) => {
        if (selected !== null) onChange(selected === "all" ? undefined : (selected as Value));
      }}
    >
      <SelectTrigger aria-label={label} className="min-w-40">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="all">{label}</SelectItem>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
