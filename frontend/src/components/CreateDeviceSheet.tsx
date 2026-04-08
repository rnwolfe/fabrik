import { useMutation, useQueryClient } from '@tanstack/react-query';
import { catalogApi } from '@/api/catalog';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet';
import DeviceForm from '@/features/catalog/DeviceForm';
import type { DeviceFormValues } from '@/features/catalog/DeviceForm';
import type { DeviceModel } from '@/models';

interface CreateDeviceSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /**
   * Called after a successful create with the newly created DeviceModel.
   * The parent can use this to auto-select the device in a picker.
   */
  onCreated?: (device: DeviceModel) => void;
}

export default function CreateDeviceSheet({
  open,
  onOpenChange,
  onCreated,
}: CreateDeviceSheetProps) {
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: catalogApi.create,
    onSuccess: (device) => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      onOpenChange(false);
      onCreated?.(device);
    },
  });

  const handleSubmit = (data: DeviceFormValues) => {
    createMutation.mutate(data);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-lg w-full overflow-y-auto">
        <SheetHeader className="pb-2">
          <SheetTitle>Create New Device</SheetTitle>
          <SheetDescription>
            Add a new device model to the catalog. It will be immediately available in
            the picker.
          </SheetDescription>
        </SheetHeader>
        <div className="px-4 pb-4">
          <DeviceForm
            onSubmit={handleSubmit}
            onCancel={handleCancel}
            isPending={createMutation.isPending}
            mutationError={createMutation.error}
            submitLabel="Create Device"
          />
        </div>
      </SheetContent>
    </Sheet>
  );
}
