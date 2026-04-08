import { useMutation, useQueryClient } from '@tanstack/react-query';
import { catalogApi, type DeviceModelRequest } from '@/api/catalog';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet';
import DeviceForm, { defaultDeviceFormValues, type DeviceFormValues } from '@/features/catalog/DeviceForm';
import type { DeviceModel } from '@/models';

interface CreateDeviceSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Called after a device is successfully created, with the new device */
  onCreated: (device: DeviceModel) => void;
}

export default function CreateDeviceSheet({
  open,
  onOpenChange,
  onCreated,
}: CreateDeviceSheetProps) {
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (data: DeviceModelRequest) => catalogApi.create(data),
    onSuccess: (newDevice) => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
      onCreated(newDevice);
      onOpenChange(false);
    },
  });

  const handleSubmit = (data: DeviceFormValues) => {
    createMutation.mutate(data);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  const handleOpenChange = (nextOpen: boolean) => {
    // Only allow closing — opening is controlled by the parent
    if (!nextOpen) {
      createMutation.reset();
      onOpenChange(false);
    }
  };

  return (
    <Sheet open={open} onOpenChange={handleOpenChange}>
      <SheetContent side="right" className="sm:max-w-lg overflow-y-auto">
        <SheetHeader>
          <SheetTitle>Create New Device</SheetTitle>
          <SheetDescription>
            Add a new device model to the catalog. It will be available immediately in
            the picker.
          </SheetDescription>
        </SheetHeader>
        <div className="px-4 pb-4">
          <DeviceForm
            initialValues={defaultDeviceFormValues}
            onSubmit={handleSubmit}
            isPending={createMutation.isPending}
            submitLabel="Create Device"
            onCancel={handleCancel}
            error={createMutation.error}
          />
        </div>
      </SheetContent>
    </Sheet>
  );
}
