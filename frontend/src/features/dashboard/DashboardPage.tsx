import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  Plus,
  Network,
  BarChart3,
  BookOpen,
  Calendar,
  FolderOpen,
  ArrowRight,
  Database,
} from 'lucide-react';
import { designsApi } from '@/api/designs';
import { useDesign } from '@/contexts/DesignContext';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import type { Design } from '@/models';

const createDesignSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  description: z.string().max(500).optional(),
});

type CreateDesignForm = z.infer<typeof createDesignSchema>;

export default function DashboardPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const { setActiveDesignId, activeDesignId } = useDesign();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: designs, isLoading } = useQuery({
    queryKey: ['designs'],
    queryFn: designsApi.list,
  });

  const createMutation = useMutation({
    mutationFn: designsApi.create,
    onSuccess: (design) => {
      queryClient.invalidateQueries({ queryKey: ['designs'] });
      setActiveDesignId(design.id);
      setCreateOpen(false);
      reset();
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateDesignForm>({
    resolver: zodResolver(createDesignSchema),
  });

  const onSubmit = (data: CreateDesignForm) => {
    createMutation.mutate(data);
  };

  const handleOpenDesign = (design: Design) => {
    setActiveDesignId(design.id);
    navigate('/design');
  };

  const formatDate = (iso: string) => {
    return new Date(iso).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <div className="mx-auto max-w-5xl">
      <PageHeader
        title="fabrik"
        subtitle="Design your datacenter network"
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="size-4" />
            New Design
          </Button>
        }
      />

      {/* Recent Designs */}
      <section className="mb-8">
        <h2 className="mb-3 text-sm font-semibold text-muted-foreground uppercase tracking-wider">
          Recent Designs
        </h2>

        {isLoading ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-32 animate-pulse rounded-xl bg-muted" />
            ))}
          </div>
        ) : !designs || designs.length === 0 ? (
          <Card>
            <CardContent>
              <EmptyState
                icon={FolderOpen}
                title="No designs yet"
                description="Create your first network topology design to get started."
                action={
                  <Button onClick={() => setCreateOpen(true)}>
                    <Plus className="size-4" />
                    Create design
                  </Button>
                }
              />
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {designs.map((design) => (
              <Card
                key={design.id}
                className={activeDesignId === design.id ? 'ring-2 ring-primary' : ''}
              >
                <CardHeader>
                  <div className="flex items-start justify-between gap-2">
                    <CardTitle className="line-clamp-1">{design.name}</CardTitle>
                    {activeDesignId === design.id && (
                      <span className="shrink-0 rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                        active
                      </span>
                    )}
                  </div>
                  {design.description && (
                    <CardDescription className="line-clamp-2">{design.description}</CardDescription>
                  )}
                </CardHeader>
                <CardFooter className="justify-between">
                  <span className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Calendar className="size-3" />
                    {formatDate(design.created_at)}
                  </span>
                  <Button size="sm" onClick={() => handleOpenDesign(design)}>
                    Open
                    <ArrowRight className="size-3" />
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </section>

      {/* Quick Actions */}
      <section>
        <h2 className="mb-3 text-sm font-semibold text-muted-foreground uppercase tracking-wider">
          Quick Actions
        </h2>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <QuickActionCard
            icon={Database}
            title="Browse Catalog"
            description="View and manage hardware device models"
            onClick={() => navigate('/catalog')}
          />
          <QuickActionCard
            icon={BarChart3}
            title="View Metrics"
            description="Analyze topology performance and capacity"
            onClick={() => navigate('/metrics')}
            disabled={!activeDesignId}
            disabledHint="Select a design first"
          />
          <QuickActionCard
            icon={BookOpen}
            title="Knowledge Base"
            description="Learn datacenter design concepts"
            onClick={() => navigate('/knowledge')}
          />
        </div>
      </section>

      {/* Network overview if active design */}
      {activeDesignId && designs && (
        <section className="mt-8">
          <div className="rounded-xl border border-border bg-muted/30 p-4">
            <div className="flex items-center gap-2 text-sm">
              <Network className="size-4 text-muted-foreground" />
              <span className="text-muted-foreground">Active design:</span>
              <span className="font-medium">
                {designs.find((d) => d.id === activeDesignId)?.name}
              </span>
              <Button
                variant="link"
                size="sm"
                className="ml-auto h-auto p-0 text-xs"
                onClick={() => navigate('/design')}
              >
                Manage fabrics →
              </Button>
            </div>
          </div>
        </section>
      )}

      {/* Create Design Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>New Design</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onSubmit)}>
            <div className="flex flex-col gap-4 py-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="e.g. Production Clos v2"
                  {...register('name')}
                  aria-invalid={!!errors.name}
                />
                {errors.name && (
                  <p className="text-xs text-destructive">{errors.name.message}</p>
                )}
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Optional description…"
                  rows={3}
                  {...register('description')}
                />
              </div>
            </div>
            <DialogFooter className="mt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCreateOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? 'Creating…' : 'Create Design'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function QuickActionCard({
  icon: Icon,
  title,
  description,
  onClick,
  disabled,
  disabledHint,
}: {
  icon: typeof Network;
  title: string;
  description: string;
  onClick: () => void;
  disabled?: boolean;
  disabledHint?: string;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      title={disabled ? disabledHint : undefined}
      className="flex items-start gap-3 rounded-xl border border-border bg-card p-4 text-left transition-colors hover:bg-muted/50 disabled:cursor-not-allowed disabled:opacity-50"
    >
      <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
        <Icon className="size-4 text-muted-foreground" />
      </div>
      <div>
        <p className="text-sm font-medium">{title}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
      </div>
    </button>
  );
}
