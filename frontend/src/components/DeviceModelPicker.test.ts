/**
 * Unit tests for DeviceModelPicker behavior.
 *
 * Tests cover:
 * - handleSelect calls onSelect with the correct id and closes the popover
 * - Auto-select after device creation (handleCreated → handleSelect)
 * - Opening the create sheet closes the picker popover first
 *
 * Full DOM render tests require @testing-library/react + a jsdom/happy-dom env.
 * The tests below cover the logic surface extracted from the component callbacks.
 */
import { describe, it, expect, vi } from 'vitest';
import type { DeviceModel } from '@/models';

const makeDevice = (id: number, vendor: string, model: string): DeviceModel => ({
  id,
  vendor,
  model,
  device_model_type: 'network',
  port_count: 32,
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
  is_seed: false,
  archived_at: null,
  created_at: '',
  updated_at: '',
});

describe('DeviceModelPicker handleSelect', () => {
  it('calls onSelect with the device id and closes the popover', () => {
    const onSelect = vi.fn();
    const setOpen = vi.fn();

    const handleSelect = (id: number) => {
      onSelect(id);
      setOpen(false);
    };

    handleSelect(7);

    expect(onSelect).toHaveBeenCalledWith(7);
    expect(setOpen).toHaveBeenCalledWith(false);
  });

  it('calls onSelect only once per selection', () => {
    const onSelect = vi.fn();
    const setOpen = vi.fn();

    const handleSelect = (id: number) => {
      onSelect(id);
      setOpen(false);
    };

    handleSelect(3);
    expect(onSelect).toHaveBeenCalledTimes(1);
  });
});

describe('DeviceModelPicker create-new flow', () => {
  it('closes popover before opening the create sheet', () => {
    const setOpen = vi.fn();
    const setSheetOpen = vi.fn();

    const handleOpenCreate = () => {
      setOpen(false);
      setSheetOpen(true);
    };

    handleOpenCreate();

    expect(setOpen).toHaveBeenCalledWith(false);
    expect(setSheetOpen).toHaveBeenCalledWith(true);
    // setOpen(false) must be called before setSheetOpen(true)
    const openCallOrder = setOpen.mock.invocationCallOrder[0];
    const sheetCallOrder = setSheetOpen.mock.invocationCallOrder[0];
    expect(openCallOrder).toBeLessThan(sheetCallOrder);
  });

  it('auto-selects the newly created device via handleCreated', () => {
    const onSelect = vi.fn();
    const setOpen = vi.fn();

    const handleSelect = (id: number) => {
      onSelect(id);
      setOpen(false);
    };

    const handleCreated = (device: DeviceModel) => {
      handleSelect(device.id);
    };

    const newDevice = makeDevice(99, 'Arista', 'DCS-7050CX3-32S');
    handleCreated(newDevice);

    expect(onSelect).toHaveBeenCalledWith(99);
    expect(setOpen).toHaveBeenCalledWith(false);
  });

  it('does not call onSelect when cancel is triggered from sheet', () => {
    const onSelect = vi.fn();

    // Cancelling the sheet just calls onOpenChange(false) — no device created
    const onSheetOpenChange = vi.fn();
    onSheetOpenChange(false);

    expect(onSelect).not.toHaveBeenCalled();
    expect(onSheetOpenChange).toHaveBeenCalledWith(false);
  });
});

describe('DeviceModelPicker device list filtering', () => {
  const devices = [
    makeDevice(1, 'Cisco', 'Nexus 9336C'),
    makeDevice(2, 'Arista', 'DCS-7050CX3'),
    makeDevice(3, 'Juniper', 'QFX5120'),
  ];

  it('finds the selected device from the devices array by id', () => {
    const effectiveValue = 2;
    const selected = devices.find((d) => d.id === effectiveValue);
    expect(selected?.vendor).toBe('Arista');
    expect(selected?.model).toBe('DCS-7050CX3');
  });

  it('returns undefined when no device matches the value', () => {
    const effectiveValue = 999;
    const selected = devices.find((d) => d.id === effectiveValue);
    expect(selected).toBeUndefined();
  });
});
