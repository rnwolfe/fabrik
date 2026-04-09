import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  Plus,
  Search,
  Pencil,
  Trash2,
  Copy,
  Database,
  ChevronDown,
} from 'lucide-react';
import { catalogApi, type DeviceModelRequest } from '@/api/catalog';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import type { DeviceModel, DeviceModelType } from '@/models';

const numField = (min = 0) => z.preprocess((v) => (v === '' || v == null ? 0 : Number(v)), z.number().min(min));

const portGroupSchema = z.object({
  count: numField(1),
  speed_gbps: numField(1),
  label: z.string().default(''),
});

const deviceSchema = z.object({
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

type DeviceForm = z.infer<typeof deviceSchema>;

const deviceTypeLabels: Record<DeviceModelType, string> = {
  network: 'Network',
  server: 'Server',
  storage: 'Storage',
  other: 'Other',
};

export default function CatalogPage() {
  const [search, setSearch] = useState('');
  const [vendorFilter, setVendorFilter] = useState<string>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editDevice, setEditDevice] = useState<DeviceModel | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<DeviceModel | null>(null);

  const queryClient = useQueryClient();

  const { data: devices, isLoading } = useQuery({
    queryKey: ['catalog'],
    queryFn: catalogApi.list,
  });

  const createMutation = useMutation({
    mutationFn: (data: DeviceModelRequest) => catalogApi.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['catalog'] }); closeDialog(); },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: DeviceModelRequest }) =>
      catalogApi.update(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['catalog'] }); closeDialog(); },
  });

  const deleteMutation = useMutation({
    mutationFn: catalogApi.delete,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['catalog'] }); setDeleteConfirm(null); },
  });

  const duplicateMutation = useMutation({
    mutationFn: catalogApi.duplicate,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['catalog'] }); },
  });

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    control,
    formState: { errors },
  } = useForm<DeviceForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(deviceSchema) as any,
    defaultValues: { device_model_type: 'network', height_u: 1, port_count: 0, port_groups: [] },
  });

  const { fields: portGroupFields, append: appendPortGroup, remove: removePortGroup } = useFieldArray({
    control,
    name: 'port_groups',
  });

  const watchedType = watch('device_model_type');
  const isNetworkType = watchedType === 'network';
  const isComputeType = watchedType === 'server' || watchedType === 'storage';

  const openCreate = () => {
    setEditDevice(null);
    reset({ device_model_type: 'network', height_u: 1, port_count: 0, port_groups: [] });
    setDialogOpen(true);
  };

  const openEdit = (device: DeviceModel) => {
    setEditDevice(device);
    reset({
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
    });
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditDevice(null);
    reset();
  };

  const onSubmit = (data: DeviceForm) => {
    if (editDevice) {
      updateMutation.mutate({ id: editDevice.id, data });
    } else {
      createMutation.mutate(data);
    }
  };

  const activeDevices = useMemo(
    () => (devices ?? []).filter((d) => !d.archived_at),
    [devices]
  );

  const vendors = useMemo(
    () => Array.from(new Set(activeDevices.map((d) => d.vendor))).sort(),
    [activeDevices]
  );

  const filtered = useMemo(() => {
    return activeDevices.filter((d) => {
      const matchSearch =
        !search ||
        d.vendor.toLowerCase().includes(search.toLowerCase()) ||
        d.model.toLowerCase().includes(search.toLowerCase());
      const matchVendor = vendorFilter === 'all' || d.vendor === vendorFilter;
      const matchType = typeFilter === 'all' || d.device_model_type === typeFilter;
      return matchSearch && matchVendor && matchType;
    });
  }, [activeDevices, search, vendorFilter, typeFilter]);

  const isPending = createMutation.isPending || updateMutation.isPending;
  const mutationError = createMutation.error ?? updateMutation.error;

  return (
    <div className="mx-auto max-w-6xl">
      <PageHeader
        title="Device Catalog"
        subtitle="Hardware models available for topology design"
        actions={
          <Button onClick={openCreate}>
            <Plus className="size-4" />
            Add Device
          </Button>
        }
      />

      {/* Filter bar */}
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[200px]">
          <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search models…"
            className="pl-8"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Select value={vendorFilter} onValueChange={(v) => setVendorFilter(v ?? 'all')}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="All vendors" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All vendors</SelectItem>
            {vendors.map((v) => (
              <SelectItem key={v} value={v}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v ?? 'all')}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="All types">
              {(value: string) => value === 'all' ? 'All types' : deviceTypeLabels[value as DeviceModelType] ?? value}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All types</SelectItem>
            <SelectItem value="network">Network</SelectItem>
            <SelectItem value="server">Server</SelectItem>
            <SelectItem value="storage">Storage</SelectItem>
            <SelectItem value="other">Other</SelectItem>
          </SelectContent>
        </Select>
        {filtered.length !== activeDevices.length && (
          <span className="text-xs text-muted-foreground">
            {filtered.length} of {activeDevices.length} devices
          </span>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-12 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={Database}
          title={search || vendorFilter !== 'all' || typeFilter !== 'all' ? 'No matching devices' : 'No devices in catalog'}
          description={
            search || vendorFilter !== 'all' || typeFilter !== 'all'
              ? 'Try adjusting your filters'
              : 'Add your first device model to get started'
          }
          action={
            !search && vendorFilter === 'all' && typeFilter === 'all' ? (
              <Button onClick={openCreate}>
                <Plus className="size-4" />
                Add device
              </Button>
            ) : undefined
          }
        />
      ) : (
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Model</TableHead>
                <TableHead>Vendor</TableHead>
                <TableHead>Type</TableHead>
                <TableHead className="text-right">Ports</TableHead>
                <TableHead className="text-right">Height (U)</TableHead>
                <TableHead className="text-right">Power (W typ)</TableHead>
                <TableHead className="w-[100px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((device) => (
                <TableRow key={device.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{device.model}</span>
                      {device.is_seed && (
                        <Badge variant="secondary" className="text-[10px]">seed</Badge>
                      )}
                    </div>
                    {device.description && (
                      <p className="mt-0.5 text-xs text-muted-foreground line-clamp-1">
                        {device.description}
                      </p>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{device.vendor}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-[10px]">
                      {deviceTypeLabels[device.device_model_type]}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {device.port_groups && device.port_groups.length > 0 ? (
                      <span title={device.port_groups.map((pg) => `${pg.count}×${pg.speed_gbps}G`).join(' + ')}>
                        {device.port_count}
                      </span>
                    ) : (
                      device.port_count
                    )}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">{device.height_u}U</TableCell>
                  <TableCell className="text-right font-mono text-sm">{device.power_watts_typical}W</TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        title="Duplicate"
                        onClick={() => duplicateMutation.mutate(device.id)}
                      >
                        <Copy className="size-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        title="Edit"
                        onClick={() => openEdit(device)}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        title="Delete"
                        onClick={() => setDeleteConfirm(device)}
                      >
                        <Trash2 className="size-3.5 text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Add/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={(o) => { if (!o) closeDialog(); }}>
        <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editDevice ? 'Edit Device' : 'Add Device'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)}>
            <div className="grid gap-4 py-2">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="vendor">Vendor</Label>
                  <Input id="vendor" placeholder="e.g. Cisco" {...register('vendor')} aria-invalid={!!errors.vendor} />
                  {errors.vendor && <p className="text-xs text-destructive">{errors.vendor.message}</p>}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="model">Model</Label>
                  <Input id="model" placeholder="e.g. Nexus 9300" {...register('model')} aria-invalid={!!errors.model} />
                  {errors.model && <p className="text-xs text-destructive">{errors.model.message}</p>}
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
                        <Input placeholder="e.g. 25GbE SFP28" {...register(`port_groups.${index}.label`)} />
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
                  <Input id="power_watts_typical" type="number" min={0} {...register('power_watts_typical')} />
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

              {mutationError && (
                <p className="text-sm text-destructive">{mutationError.message}</p>
              )}
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeDialog}>Cancel</Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Saving…' : editDevice ? 'Save Changes' : 'Add Device'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={!!deleteConfirm} onOpenChange={(o) => { if (!o) setDeleteConfirm(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Device?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            This will archive <strong>{deleteConfirm?.vendor} {deleteConfirm?.model}</strong>. It will no longer appear in the catalog.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirm(null)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => deleteConfirm && deleteMutation.mutate(deleteConfirm.id)}
            >
              {deleteMutation.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
