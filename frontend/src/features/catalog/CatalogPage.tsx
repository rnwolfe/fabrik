import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Plus,
  Search,
  Pencil,
  Trash2,
  Copy,
  Database,
} from 'lucide-react';
import { catalogApi } from '@/api/catalog';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
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
import DeviceForm, { deviceTypeLabels } from './DeviceForm';
import type { DeviceFormValues } from './DeviceForm';

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
    mutationFn: catalogApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      closeDialog();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: DeviceFormValues }) =>
      catalogApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      closeDialog();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: catalogApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      setDeleteConfirm(null);
    },
  });

  const duplicateMutation = useMutation({
    mutationFn: catalogApi.duplicate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
    },
  });

  const openCreate = () => {
    setEditDevice(null);
    setDialogOpen(true);
  };

  const openEdit = (device: DeviceModel) => {
    setEditDevice(device);
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditDevice(null);
  };

  const handleFormSubmit = (data: DeviceFormValues) => {
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
          <DeviceForm
            editDevice={editDevice}
            onSubmit={handleFormSubmit}
            onCancel={closeDialog}
            isPending={isPending}
            mutationError={mutationError}
          />
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
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setDeleteConfirm(null)}>Cancel</Button>
            <Button
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => deleteConfirm && deleteMutation.mutate(deleteConfirm.id)}
            >
              {deleteMutation.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
