import { useState, useCallback, createContext, useContext } from 'react';
import { Popover as PopoverPrimitive } from '@base-ui/react/popover';
import { ChevronsUpDown, PlusCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command';
import type { DeviceModel } from '@/models';
import CreateDeviceSheet from './CreateDeviceSheet';
import { suggestedRoles, roleBadgeLabel, roleBadgeVariant, type Role } from '@/lib/deviceRoles';
import { portGroupSummary } from '@/lib/portGroups';

// ── Nested popover guard ──────────────────────────────────────────────────────
// When a DeviceModelPicker popover is open inside a Dialog, clicks on the
// popover's portaled content register as "outside" the dialog and dismiss it.
// PickerGuardProvider / usePickerGuard let the Dialog suppress dismiss while
// any picker popover is open.

interface PickerGuardCtx {
  pickerOpen: boolean;
  setPickerOpen: (v: boolean) => void;
}
const PickerGuardContext = createContext<PickerGuardCtx | null>(null);

/**
 * Wrap a Dialog that contains a DeviceModelPicker with this provider.
 * Then use `usePickerGuard()` in the dialog's `onOpenChange` to suppress
 * dismissals while the picker is open.
 */
export function PickerGuardProvider({ children }: { children: React.ReactNode }) {
  const [pickerOpen, setPickerOpen] = useState(false);
  return (
    <PickerGuardContext.Provider value={{ pickerOpen, setPickerOpen }}>
      {children}
    </PickerGuardContext.Provider>
  );
}

export function usePickerGuard() {
  const ctx = useContext(PickerGuardContext);
  return ctx?.pickerOpen ?? false;
}

// ── Role badge ─────────────────────────────────────────────────────────────────

function RoleBadge({ role }: { role: Role }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded px-1 py-0.5 text-[9px] font-medium leading-none',
        roleBadgeVariant[role],
      )}
    >
      {roleBadgeLabel[role]}
    </span>
  );
}

// ── Component ─────────────────────────────────────────────────────────────────

interface DeviceModelPickerProps {
  devices: DeviceModel[];
  value?: number;
  onSelect: (deviceModelId: number) => void;
  placeholder?: string;
  className?: string;
  triggerClassName?: string;
  /** When provided, devices that don't match the role are de-emphasized (grayed, pushed to bottom) but remain selectable. */
  role?: Role;
}

export default function DeviceModelPicker({
  devices,
  value,
  onSelect,
  placeholder = 'Select device model…',
  className,
  triggerClassName,
  role,
}: DeviceModelPickerProps) {
  const [open, setOpenRaw] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);
  const guard = useContext(PickerGuardContext);

  // Support both controlled (value prop) and uncontrolled (internal) usage
  const [internalValue, setInternalValue] = useState<number | undefined>(undefined);
  const effectiveValue = value ?? internalValue;
  const selected = effectiveValue ? devices.find((d) => d.id === effectiveValue) : undefined;

  const setOpen = useCallback(
    (v: boolean) => {
      setOpenRaw(v);
      guard?.setPickerOpen(v);
    },
    [guard],
  );

  const handleSelect = useCallback(
    (id: number) => {
      setInternalValue(id);
      onSelect(id);
      setOpen(false);
    },
    [onSelect, setOpen],
  );

  const handleOpenCreate = useCallback(() => {
    // Close the popover before opening the sheet
    setOpen(false);
    setSheetOpen(true);
  }, [setOpen]);

  const handleCreated = useCallback(
    (device: DeviceModel) => {
      // Auto-select the newly created device
      handleSelect(device.id);
    },
    [handleSelect],
  );

  // Sort: role-matching devices first, then non-matching.
  const sortedDevices = role
    ? [
        ...devices.filter((d) => suggestedRoles(d).includes(role)),
        ...devices.filter((d) => !suggestedRoles(d).includes(role)),
      ]
    : devices;

  return (
    <>
      <PopoverPrimitive.Root open={open} onOpenChange={setOpen}>
        <PopoverPrimitive.Trigger
          render={
            <Button
              variant="outline"
              className={cn(
                'w-full justify-between font-normal',
                !selected && 'text-muted-foreground',
                triggerClassName,
              )}
            />
          }
        >
          {selected ? (
            <span className="truncate">
              {selected.vendor} {selected.model}
            </span>
          ) : (
            <span>{placeholder}</span>
          )}
          <ChevronsUpDown className="ml-2 size-3.5 shrink-0 opacity-50" />
        </PopoverPrimitive.Trigger>
        <PopoverPrimitive.Portal>
          <PopoverPrimitive.Positioner align="start" sideOffset={4} className="isolate z-[100]">
            <PopoverPrimitive.Popup
              className={cn(
                'w-80 rounded-lg bg-popover p-0 text-sm text-popover-foreground shadow-md ring-1 ring-foreground/10 outline-hidden origin-(--transform-origin) duration-100 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95',
                className,
              )}
            >
              <Command>
                <CommandInput placeholder="Search devices…" />
                <CommandList>
                  <CommandEmpty>No devices found.</CommandEmpty>
                  <CommandGroup>
                    {sortedDevices.map((dm) => {
                      const dmRoles = suggestedRoles(dm);
                      const isDeemphasized = role !== undefined && !dmRoles.includes(role);
                      return (
                        <CommandItem
                          key={dm.id}
                          value={`${dm.vendor} ${dm.model} ${dm.description}`}
                          onSelect={() => handleSelect(dm.id)}
                          data-checked={effectiveValue === dm.id ? 'true' : undefined}
                          className={cn(isDeemphasized && 'opacity-40')}
                        >
                          <div className="flex flex-col gap-0.5 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium truncate">
                                {dm.vendor} {dm.model}
                              </span>
                              <div className="flex items-center gap-1 ml-auto shrink-0">
                                {dmRoles.map((r) => (
                                  <RoleBadge key={r} role={r} />
                                ))}
                              </div>
                            </div>
                            <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                              {portGroupSummary(dm) !== '—' && <span>{portGroupSummary(dm)}</span>}
                              {portGroupSummary(dm) === '—' && dm.port_count > 0 && <span>{dm.port_count} ports</span>}
                              {dm.power_watts_typical > 0 && (
                                <>
                                  <span className="text-border">·</span>
                                  <span>{dm.power_watts_typical}W</span>
                                </>
                              )}
                              {dm.height_u > 0 && (
                                <>
                                  <span className="text-border">·</span>
                                  <span>{dm.height_u}U</span>
                                </>
                              )}
                            </div>
                            {dm.description && (
                              <p className="text-[10px] text-muted-foreground/70 truncate">
                                {dm.description}
                              </p>
                            )}
                          </div>
                        </CommandItem>
                      );
                    })}
                  </CommandGroup>
                  <CommandSeparator />
                  <CommandGroup>
                    <CommandItem
                      value="__create_new_device__"
                      onSelect={handleOpenCreate}
                      className="text-muted-foreground"
                    >
                      <PlusCircle className="mr-2 size-3.5" />
                      Create new device…
                    </CommandItem>
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverPrimitive.Popup>
          </PopoverPrimitive.Positioner>
        </PopoverPrimitive.Portal>
      </PopoverPrimitive.Root>

      <CreateDeviceSheet
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        onCreated={handleCreated}
      />
    </>
  );
}
