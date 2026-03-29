import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  Plus,
  Pencil,
  Trash2,
  Rows3,
  ChevronDown,
  ChevronRight,
  Server,
} from 'lucide-react';
import { racksApi } from '@/api/racks';
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
import type { RackType, RackSummary } from '@/models';

const rackTypeSchema = z.object({
  name: z.string().min(1, 'Name required'),
  height_u: z.preprocess((v) => (v === '' || v == null ? 42 : Number(v)), z.number().int().min(1)),
  power_capacity_w: z.preprocess((v) => (v === '' || v == null ? 10000 : Number(v)), z.number().min(0)),
  description: z.string().default(''),
});
type RackTypeForm = z.infer<typeof rackTypeSchema>;

const rackSchema = z.object({
  name: z.string().min(1, 'Name required'),
  rack_type_id: z.preprocess(
    (v) => (v === '' || v == null ? undefined : Number(v)),
    z.number().int().optional()
  ),
  description: z.string().default(''),
});
type RackForm = z.infer<typeof rackSchema>;

export default function RacksPage() {
  const [rackTypeDialogOpen, setRackTypeDialogOpen] = useState(false);
  const [editRackType, setEditRackType] = useState<RackType | null>(null);
  const [deleteRackType, setDeleteRackType] = useState<RackType | null>(null);

  const [rackDialogOpen, setRackDialogOpen] = useState(false);
  const [editRack, setEditRack] = useState<RackSummary | null>(null);
  const [deleteRack, setDeleteRack] = useState<RackSummary | null>(null);
  const [expandedRack, setExpandedRack] = useState<number | null>(null);

  const queryClient = useQueryClient();

  const { data: rackTypes, isLoading: typesLoading } = useQuery({
    queryKey: ['rack-types'],
    queryFn: racksApi.listTypes,
  });

  const { data: racks, isLoading: racksLoading } = useQuery({
    queryKey: ['racks'],
    queryFn: racksApi.list,
  });

  // Rack type mutations
  const createTypeMutation = useMutation({
    mutationFn: (d: Partial<RackType>) => racksApi.createType(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['rack-types'] }); setRackTypeDialogOpen(false); },
  });
  const updateTypeMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<RackType> }) => racksApi.updateType(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['rack-types'] }); closeTypeDialog(); },
  });
  const deleteTypeMutation = useMutation({
    mutationFn: racksApi.deleteType,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['rack-types'] }); setDeleteRackType(null); },
  });

  // Rack mutations
  const createRackMutation = useMutation({
    mutationFn: (d: Partial<RackSummary>) => racksApi.create(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['racks'] }); setRackDialogOpen(false); },
  });
  const updateRackMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<RackSummary> }) => racksApi.update(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['racks'] }); closeRackDialog(); },
  });
  const deleteRackMutation = useMutation({
    mutationFn: racksApi.delete,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['racks'] }); setDeleteRack(null); },
  });

  // Rack type form
  const {
    register: registerType,
    handleSubmit: handleSubmitType,
    reset: resetType,
    formState: { errors: typeErrors },
  } = useForm<RackTypeForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(rackTypeSchema) as any,
    defaultValues: { height_u: 42, power_capacity_w: 10000 },
  });

  // Rack form
  const {
    register: registerRack,
    handleSubmit: handleSubmitRack,
    reset: resetRack,
    setValue: setRackValue,
    watch: watchRack,
    formState: { errors: rackErrors },
  } = useForm<RackForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(rackSchema) as any,
  });

  const watchedRackTypeId = watchRack('rack_type_id');

  const openTypeCreate = () => {
    setEditRackType(null);
    resetType({ height_u: 42, power_capacity_w: 10000 });
    setRackTypeDialogOpen(true);
  };
  const openTypeEdit = (rt: RackType) => {
    setEditRackType(rt);
    resetType({ name: rt.name, height_u: rt.height_u, power_capacity_w: rt.power_capacity_w, description: rt.description });
    setRackTypeDialogOpen(true);
  };
  const closeTypeDialog = () => { setRackTypeDialogOpen(false); setEditRackType(null); };

  const onSubmitType = (data: RackTypeForm) => {
    if (editRackType) {
      updateTypeMutation.mutate({ id: editRackType.id, data });
    } else {
      createTypeMutation.mutate(data);
    }
  };

  const openRackCreate = () => {
    setEditRack(null);
    resetRack({ description: '' });
    setRackDialogOpen(true);
  };
  const openRackEdit = (rack: RackSummary) => {
    setEditRack(rack);
    resetRack({ name: rack.name, description: rack.description });
    setRackDialogOpen(true);
  };
  const closeRackDialog = () => { setRackDialogOpen(false); setEditRack(null); };

  const onSubmitRack = (data: RackForm) => {
    if (editRack) {
      updateRackMutation.mutate({ id: editRack.id, data });
    } else {
      createRackMutation.mutate(data);
    }
  };

  const typeIsPending = createTypeMutation.isPending || updateTypeMutation.isPending;
  const rackIsPending = createRackMutation.isPending || updateRackMutation.isPending;

  return (
    <div className="mx-auto max-w-5xl space-y-8">
      <PageHeader
        title="Racks"
        subtitle="Rack types and physical rack inventory"
        actions={
          <Button onClick={openRackCreate}>
            <Plus className="size-4" />
            Add Rack
          </Button>
        }
      />

      {/* Rack Types Section */}
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
            Rack Types
          </h2>
          <Button variant="outline" size="sm" onClick={openTypeCreate}>
            <Plus className="size-3.5" />
            Add Type
          </Button>
        </div>

        {typesLoading ? (
          <div className="h-24 animate-pulse rounded-xl bg-muted" />
        ) : !rackTypes || rackTypes.length === 0 ? (
          <EmptyState
            icon={Rows3}
            title="No rack types"
            description="Define rack types to use as templates when creating racks."
            action={
              <Button variant="outline" size="sm" onClick={openTypeCreate}>
                <Plus className="size-3.5" />
                Add type
              </Button>
            }
          />
        ) : (
          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead className="text-right">Height (U)</TableHead>
                  <TableHead className="text-right">Power (W)</TableHead>
                  <TableHead className="w-[80px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rackTypes.map((rt) => (
                  <TableRow key={rt.id}>
                    <TableCell>
                      <p className="font-medium">{rt.name}</p>
                      {rt.description && (
                        <p className="text-xs text-muted-foreground">{rt.description}</p>
                      )}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm">{rt.height_u}U</TableCell>
                    <TableCell className="text-right font-mono text-sm">{rt.power_capacity_w.toLocaleString()}W</TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end gap-1">
                        <Button variant="ghost" size="icon-sm" onClick={() => openTypeEdit(rt)}>
                          <Pencil className="size-3.5" />
                        </Button>
                        <Button variant="ghost" size="icon-sm" onClick={() => setDeleteRackType(rt)}>
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
      </section>

      {/* Racks Section */}
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
            Racks
          </h2>
        </div>

        {racksLoading ? (
          <div className="h-32 animate-pulse rounded-xl bg-muted" />
        ) : !racks || racks.length === 0 ? (
          <EmptyState
            icon={Server}
            title="No racks"
            description="Add racks to your datacenter design."
            action={
              <Button onClick={openRackCreate}>
                <Plus className="size-4" />
                Add rack
              </Button>
            }
          />
        ) : (
          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead className="text-right">Used / Total (U)</TableHead>
                  <TableHead className="text-right">Power Draw (W)</TableHead>
                  <TableHead className="text-right">Devices</TableHead>
                  <TableHead className="w-[100px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {racks.map((rack) => (
                  <>
                    <TableRow key={rack.id}>
                      <TableCell>
                        <div className="flex items-center gap-1.5">
                          <button
                            onClick={() => setExpandedRack(expandedRack === rack.id ? null : rack.id)}
                            className="flex items-center gap-1.5 hover:text-foreground text-muted-foreground"
                          >
                            {expandedRack === rack.id ? (
                              <ChevronDown className="size-3.5" />
                            ) : (
                              <ChevronRight className="size-3.5" />
                            )}
                          </button>
                          <span className="font-medium">{rack.name}</span>
                          {rack.warning && (
                            <Badge variant="destructive" className="text-[10px]">warning</Badge>
                          )}
                        </div>
                        {rack.description && (
                          <p className="ml-5 text-xs text-muted-foreground">{rack.description}</p>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        <span className={rack.used_u > rack.height_u ? 'text-destructive' : ''}>
                          {rack.used_u}
                        </span>
                        <span className="text-muted-foreground"> / {rack.height_u}U</span>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {rack.used_watts_typical.toLocaleString()}W
                        <span className="ml-1 text-xs text-muted-foreground">
                          / {rack.power_capacity_w.toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {rack.devices?.length ?? 0}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="icon-sm" onClick={() => openRackEdit(rack)}>
                            <Pencil className="size-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon-sm" onClick={() => setDeleteRack(rack)}>
                            <Trash2 className="size-3.5 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                    {expandedRack === rack.id && rack.devices && rack.devices.length > 0 && (
                      <TableRow key={`${rack.id}-devices`}>
                        <TableCell colSpan={5} className="p-0">
                          <div className="bg-muted/30 px-6 py-2">
                            <p className="mb-1 text-xs font-medium text-muted-foreground">Installed Devices</p>
                            <div className="space-y-0.5">
                              {rack.devices.map((device) => (
                                <div key={device.id} className="flex items-center gap-3 text-xs">
                                  <span className="w-8 text-right font-mono text-muted-foreground">{device.position}U</span>
                                  <span className="font-medium">{device.name}</span>
                                  <span className="text-muted-foreground">{device.model_vendor} {device.model_name}</span>
                                  <Badge variant="outline" className="text-[10px]">{device.role}</Badge>
                                  <span className="ml-auto text-muted-foreground">{device.height_u}U · {device.power_watts_typical}W</span>
                                </div>
                              ))}
                            </div>
                          </div>
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </section>

      {/* Rack Type Dialog */}
      <Dialog open={rackTypeDialogOpen} onOpenChange={(o) => { if (!o) closeTypeDialog(); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>{editRackType ? 'Edit Rack Type' : 'Add Rack Type'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmitType(onSubmitType)}>
            <div className="flex flex-col gap-4 py-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rt-name">Name</Label>
                <Input id="rt-name" placeholder="e.g. Standard 42U" {...registerType('name')} aria-invalid={!!typeErrors.name} />
                {typeErrors.name && <p className="text-xs text-destructive">{typeErrors.name.message}</p>}
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="rt-height">Height (U)</Label>
                  <Input id="rt-height" type="number" min={1} {...registerType('height_u')} />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="rt-power">Power (W)</Label>
                  <Input id="rt-power" type="number" min={0} {...registerType('power_capacity_w')} />
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rt-desc">Description</Label>
                <Textarea id="rt-desc" rows={2} {...registerType('description')} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeTypeDialog}>Cancel</Button>
              <Button type="submit" disabled={typeIsPending}>
                {typeIsPending ? 'Saving…' : editRackType ? 'Save' : 'Add Type'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Rack Dialog */}
      <Dialog open={rackDialogOpen} onOpenChange={(o) => { if (!o) closeRackDialog(); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>{editRack ? 'Edit Rack' : 'Add Rack'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmitRack(onSubmitRack)}>
            <div className="flex flex-col gap-4 py-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rack-name">Name</Label>
                <Input id="rack-name" placeholder="e.g. Row A, Rack 01" {...registerRack('name')} aria-invalid={!!rackErrors.name} />
                {rackErrors.name && <p className="text-xs text-destructive">{rackErrors.name.message}</p>}
              </div>
              {rackTypes && rackTypes.length > 0 && (
                <div className="flex flex-col gap-1.5">
                  <Label>Rack Type (optional)</Label>
                  <Select
                    value={watchedRackTypeId ? String(watchedRackTypeId) : ''}
                    onValueChange={(v) => setRackValue('rack_type_id', (v && v !== '') ? Number(v) : undefined)}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="No type selected">
                        {(value: string) => {
                          if (!value) return 'No type selected';
                          const rt = rackTypes.find((t) => String(t.id) === value);
                          return rt ? `${rt.name} (${rt.height_u}U)` : 'No type selected';
                        }}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">None</SelectItem>
                      {rackTypes.map((rt) => (
                        <SelectItem key={rt.id} value={String(rt.id)}>
                          {rt.name} ({rt.height_u}U)
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rack-desc">Description</Label>
                <Textarea id="rack-desc" rows={2} {...registerRack('description')} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeRackDialog}>Cancel</Button>
              <Button type="submit" disabled={rackIsPending}>
                {rackIsPending ? 'Saving…' : editRack ? 'Save' : 'Add Rack'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete rack type confirm */}
      <Dialog open={!!deleteRackType} onOpenChange={(o) => { if (!o) setDeleteRackType(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader><DialogTitle>Delete Rack Type?</DialogTitle></DialogHeader>
          <p className="text-sm text-muted-foreground">
            Delete rack type <strong>{deleteRackType?.name}</strong>?
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteRackType(null)}>Cancel</Button>
            <Button variant="destructive" disabled={deleteTypeMutation.isPending}
              onClick={() => deleteRackType && deleteTypeMutation.mutate(deleteRackType.id)}>
              {deleteTypeMutation.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete rack confirm */}
      <Dialog open={!!deleteRack} onOpenChange={(o) => { if (!o) setDeleteRack(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader><DialogTitle>Delete Rack?</DialogTitle></DialogHeader>
          <p className="text-sm text-muted-foreground">
            Delete rack <strong>{deleteRack?.name}</strong> and all its devices?
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteRack(null)}>Cancel</Button>
            <Button variant="destructive" disabled={deleteRackMutation.isPending}
              onClick={() => deleteRack && deleteRackMutation.mutate(deleteRack.id)}>
              {deleteRackMutation.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
