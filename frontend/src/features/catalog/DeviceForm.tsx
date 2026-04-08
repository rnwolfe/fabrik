import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  Plus,
  Trash2,
  ChevronDown,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { DeviceModel, DeviceModelType } from '@/models';

// ── Schema ────────────────────────────────────────────────────────────────────

const numField = (min = 0) =>
  z.preprocess((v) => (v === '' || v == null ? 0 : Number(v)), z.number().min(min));

const portGroupSchema = z.object({
  count: numField(1),
  speed_gbps: numField(1),
  label: z.string().default(''),
});

export const deviceFormSchema = z.object({
  vendor: z.string().min(1, 'Vendor required'),
  model: z.string().min(1, 'Model required'),
  device_model_type: z.enum(['network', 'server', 'storage', 'other']),
  port_count: numField(0),
  port_groups: z.array(portGroupSchema).default([]),
  height_u: numField(1),
  power_watts_idle: numField(0),
  power_watts_typical: numField(0),
  power_watts_max: numField(0),
  cpu_sockets: numField(0),
  cores_per_socket: numField(0),
  ram_gb: numField(0),
  storage_tb: numField(0),
  gpu_count: numField(0),
  description: z.string().default(''),
});

export type DeviceFormValues = z.infer<typeof deviceFormSchema>;

export const deviceTypeLabels: Record<DeviceModelType, string> = {
  network: 'Network',
  server: 'Server',
  storage: 'Storage',
  other: 'Other',
};

export const defaultDeviceFormValues: Partial<DeviceFormValues> = {
  device_model_type: 'network',
  height_u: 1,
  port_count: 0,
  port_groups: [],
};

export function deviceToFormValues(device: DeviceModel): DeviceFormValues {
  return {
    vendor: device.vendor,
    model: device.model,
    device_model_type: device.device_model_type,
    port_count: device.port_count,
    port_groups: (device.port_groups ?? []).map((pg) => ({
      count: pg.count,
      speed_gbps: pg.speed_gbps,
      label: pg.label,
    })),
    height_u: device.height_u,
    power_watts_idle: device.power_watts_idle,
    power_watts_typical: device.power_watts_typical,
    power_watts_max: device.power_watts_max,
    cpu_sockets: device.cpu_sockets,
    cores_per_socket: device.cores_per_socket,
    ram_gb: device.ram_gb,
    storage_tb: device.storage_tb,
    gpu_count: device.gpu_count,
    description: device.description,
  };
}

// ── Component ─────────────────────────────────────────────────────────────────

interface DeviceFormProps {
  /** Existing device being edited. If undefined, this is a create form. */
  editDevice?: DeviceModel | null;
  /** Called with validated form values on submit. */
  onSubmit: (data: DeviceFormValues) => void;
  /** Called when the user clicks Cancel. */
  onCancel: () => void;
  /** Whether the form submission is in progress. */
  isPending?: boolean;
  /** Server-side / mutation error to show below the form. */
  mutationError?: Error | null;
  /** Optional submit button label override. */
  submitLabel?: string;
}

export default function DeviceForm({
  editDevice,
  onSubmit,
  onCancel,
  isPending,
  mutationError,
  submitLabel,
}: DeviceFormProps) {
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    control,
    formState: { errors },
  } = useForm<DeviceFormValues>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(deviceFormSchema) as any,
    defaultValues: editDevice ? deviceToFormValues(editDevice) : defaultDeviceFormValues,
  });

  const { fields: portGroupFields, append: appendPortGroup, remove: removePortGroup } =
    useFieldArray({ control, name: 'port_groups' });

  const watchedType = watch('device_model_type');
  const isNetworkType = watchedType === 'network';
  const isComputeType = watchedType === 'server' || watchedType === 'storage';

  const resolvedSubmitLabel =
    submitLabel ?? (editDevice ? 'Save Changes' : 'Add Device');

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="grid gap-4 py-2">
        {/* Vendor + Model */}
        <div className="grid grid-cols-2 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-vendor">Vendor</Label>
            <Input
              id="df-vendor"
              placeholder="e.g. Cisco"
              {...register('vendor')}
              aria-invalid={!!errors.vendor}
            />
            {errors.vendor && (
              <p className="text-xs text-destructive">{errors.vendor.message}</p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-model">Model</Label>
            <Input
              id="df-model"
              placeholder="e.g. Nexus 9300"
              {...register('model')}
              aria-invalid={!!errors.model}
            />
            {errors.model && (
              <p className="text-xs text-destructive">{errors.model.message}</p>
            )}
          </div>
        </div>

        {/* Device Type */}
        <div className="flex flex-col gap-1.5">
          <Label>Device Type</Label>
          <Select
            value={watchedType}
            onValueChange={(v) => setValue('device_model_type', v as DeviceModelType)}
          >
            <SelectTrigger className="w-full">
              <SelectValue>
                {(value: string) => deviceTypeLabels[value as DeviceModelType] ?? value}
              </SelectValue>
              <ChevronDown className="size-4" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="network">Network</SelectItem>
              <SelectItem value="server">Server</SelectItem>
              <SelectItem value="storage">Storage</SelectItem>
              <SelectItem value="other">Other</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Port Count + Height */}
        <div className="grid grid-cols-3 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-port-count">Port Count</Label>
            <Input id="df-port-count" type="number" min={0} {...register('port_count')} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-height-u">Height (U)</Label>
            <Input id="df-height-u" type="number" min={1} {...register('height_u')} />
          </div>
        </div>

        {/* Port Groups (network only) */}
        {isNetworkType && (
          <div className="space-y-3 border-t pt-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Port Groups
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Physical port sets available on this device
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => appendPortGroup({ count: 0, speed_gbps: 0, label: '' })}
              >
                <Plus className="size-3" />
                Add Group
              </Button>
            </div>
            {portGroupFields.map((field, index) => (
              <div
                key={field.id}
                className="grid grid-cols-[1fr_1fr_1.5fr_auto] gap-2 items-end"
              >
                <div className="flex flex-col gap-1">
                  <Label className="text-[10px]">Count</Label>
                  <Input
                    type="number"
                    min={1}
                    {...register(`port_groups.${index}.count`)}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-[10px]">Speed (Gbps)</Label>
                  <Input
                    type="number"
                    min={1}
                    {...register(`port_groups.${index}.speed_gbps`)}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-[10px]">Label</Label>
                  <Input
                    placeholder="e.g. 25GbE SFP28"
                    {...register(`port_groups.${index}.label`)}
                  />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => removePortGroup(index)}
                >
                  <Trash2 className="size-3.5 text-destructive" />
                </Button>
              </div>
            ))}
            {portGroupFields.length === 0 && (
              <p className="text-xs text-muted-foreground text-center py-2">
                No port groups defined. Add groups to enable speed-based oversubscription.
              </p>
            )}
          </div>
        )}

        {/* Power */}
        <div className="grid grid-cols-3 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-power-idle">Power Idle (W)</Label>
            <Input
              id="df-power-idle"
              type="number"
              min={0}
              {...register('power_watts_idle')}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-power-typical">Power Typical (W)</Label>
            <Input
              id="df-power-typical"
              type="number"
              min={0}
              {...register('power_watts_typical')}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="df-power-max">Power Max (W)</Label>
            <Input
              id="df-power-max"
              type="number"
              min={0}
              {...register('power_watts_max')}
            />
          </div>
        </div>

        {/* Compute Specs (server/storage only) */}
        {isComputeType && (
          <div className="grid grid-cols-2 gap-4 border-t pt-4">
            <div className="col-span-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Compute Specs
              </p>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="df-cpu-sockets">CPU Sockets</Label>
              <Input id="df-cpu-sockets" type="number" min={0} {...register('cpu_sockets')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="df-cores-per-socket">Cores / Socket</Label>
              <Input
                id="df-cores-per-socket"
                type="number"
                min={0}
                {...register('cores_per_socket')}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="df-ram-gb">RAM (GB)</Label>
              <Input id="df-ram-gb" type="number" min={0} {...register('ram_gb')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="df-storage-tb">Storage (TB)</Label>
              <Input
                id="df-storage-tb"
                type="number"
                min={0}
                step="0.1"
                {...register('storage_tb')}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="df-gpu-count">GPU Count</Label>
              <Input id="df-gpu-count" type="number" min={0} {...register('gpu_count')} />
            </div>
          </div>
        )}

        {/* Description */}
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="df-description">Description</Label>
          <Textarea
            id="df-description"
            rows={2}
            placeholder="Optional…"
            {...register('description')}
          />
        </div>

        {/* Server-side error */}
        {mutationError && (
          <p className="text-sm text-destructive">{mutationError.message}</p>
        )}
      </div>

      {/* Footer actions */}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? 'Saving…' : resolvedSubmitLabel}
        </Button>
      </div>
    </form>
  );
}
