import { useEffect } from 'react';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Plus, Trash2, ChevronDown } from 'lucide-react';
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

export const numField = (min = 0) =>
  z.preprocess((v) => (v === '' || v == null ? 0 : Number(v)), z.number().min(min));

export const portGroupSchema = z.object({
  count: numField(1),
  speed_gbps: numField(1),
  label: z.string().default(''),
});

export const deviceSchema = z.object({
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

export type DeviceFormValues = z.infer<typeof deviceSchema>;

export const deviceTypeLabels: Record<DeviceModelType, string> = {
  network: 'Network',
  server: 'Server',
  storage: 'Storage',
  other: 'Other',
};

export const defaultDeviceFormValues: DeviceFormValues = {
  vendor: '',
  model: '',
  device_model_type: 'network',
  port_count: 0,
  port_groups: [],
  height_u: 1,
  power_watts_idle: 0,
  power_watts_typical: 0,
  power_watts_max: 0,
  cpu_sockets: 0,
  cores_per_socket: 0,
  ram_gb: 0,
  storage_tb: 0,
  gpu_count: 0,
  description: '',
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

interface DeviceFormProps {
  /** If provided, seeds the form with the device's current values for editing */
  initialValues?: DeviceFormValues;
  onSubmit: (data: DeviceFormValues) => void;
  isPending: boolean;
  submitLabel?: string;
  onCancel: () => void;
  error?: Error | null;
}

export default function DeviceForm({
  initialValues,
  onSubmit,
  isPending,
  submitLabel = 'Add Device',
  onCancel,
  error,
}: DeviceFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    control,
    formState: { errors },
  } = useForm<DeviceFormValues>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(deviceSchema) as any,
    defaultValues: initialValues ?? defaultDeviceFormValues,
  });

  // Sync form values when initialValues changes (e.g., switching from create to edit)
  useEffect(() => {
    if (initialValues) {
      reset(initialValues);
    }
  }, [initialValues, reset]);

  const { fields: portGroupFields, append: appendPortGroup, remove: removePortGroup } = useFieldArray({
    control,
    name: 'port_groups',
  });

  const watchedType = watch('device_model_type');
  const isNetworkType = watchedType === 'network';
  const isComputeType = watchedType === 'server' || watchedType === 'storage';

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="grid gap-4 py-2">
        <div className="grid grid-cols-2 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="vendor">Vendor</Label>
            <Input
              id="vendor"
              placeholder="e.g. Cisco"
              {...register('vendor')}
              aria-invalid={!!errors.vendor}
            />
            {errors.vendor && (
              <p className="text-xs text-destructive">{errors.vendor.message}</p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="model">Model</Label>
            <Input
              id="model"
              placeholder="e.g. Nexus 9300"
              {...register('model')}
              aria-invalid={!!errors.model}
            />
            {errors.model && (
              <p className="text-xs text-destructive">{errors.model.message}</p>
            )}
          </div>
        </div>

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

        <div className="grid grid-cols-3 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="port_count">Port Count</Label>
            <Input id="port_count" type="number" min={0} {...register('port_count')} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="height_u">Height (U)</Label>
            <Input id="height_u" type="number" min={1} {...register('height_u')} />
          </div>
        </div>

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
              <div key={field.id} className="grid grid-cols-[1fr_1fr_1.5fr_auto] gap-2 items-end">
                <div className="flex flex-col gap-1">
                  <Label className="text-[10px]">Count</Label>
                  <Input type="number" min={1} {...register(`port_groups.${index}.count`)} />
                </div>
                <div className="flex flex-col gap-1">
                  <Label className="text-[10px]">Speed (Gbps)</Label>
                  <Input type="number" min={1} {...register(`port_groups.${index}.speed_gbps`)} />
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

        <div className="grid grid-cols-3 gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="power_watts_idle">Power Idle (W)</Label>
            <Input id="power_watts_idle" type="number" min={0} {...register('power_watts_idle')} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="power_watts_typical">Power Typical (W)</Label>
            <Input
              id="power_watts_typical"
              type="number"
              min={0}
              {...register('power_watts_typical')}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="power_watts_max">Power Max (W)</Label>
            <Input id="power_watts_max" type="number" min={0} {...register('power_watts_max')} />
          </div>
        </div>

        {isComputeType && (
          <div className="grid grid-cols-2 gap-4 border-t pt-4">
            <div className="col-span-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Compute Specs
              </p>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="cpu_sockets">CPU Sockets</Label>
              <Input id="cpu_sockets" type="number" min={0} {...register('cpu_sockets')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="cores_per_socket">Cores / Socket</Label>
              <Input id="cores_per_socket" type="number" min={0} {...register('cores_per_socket')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="ram_gb">RAM (GB)</Label>
              <Input id="ram_gb" type="number" min={0} {...register('ram_gb')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="storage_tb">Storage (TB)</Label>
              <Input id="storage_tb" type="number" min={0} step="0.1" {...register('storage_tb')} />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="gpu_count">GPU Count</Label>
              <Input id="gpu_count" type="number" min={0} {...register('gpu_count')} />
            </div>
          </div>
        )}

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="description">Description</Label>
          <Textarea id="description" rows={2} placeholder="Optional…" {...register('description')} />
        </div>

        {error && <p className="text-sm text-destructive">{error.message}</p>}
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? 'Saving…' : submitLabel}
        </Button>
      </div>
    </form>
  );
}
