import { describe, it, expect } from 'vitest';
import { errorToCopy } from './errorCopy';
import { ApiError } from '@/api/client';

describe('errorToCopy', () => {
  it('maps known code "not_found" to specific copy', () => {
    const err = new ApiError('not_found', 'not found', [], 404);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Not found');
    expect(copy.description).toBe('The requested item no longer exists.');
  });

  it('maps known code "constraint_violation" to specific copy', () => {
    const err = new ApiError('constraint_violation', 'name is required', [], 422);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Constraint violation');
    expect(copy.description).toBe('name is required');
  });

  it('maps known code "duplicate" to specific copy', () => {
    const err = new ApiError('duplicate', 'duplicate', [], 409);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Already exists');
  });

  it('maps known code "seed_read_only" to specific copy', () => {
    const err = new ApiError('seed_read_only', 'seed device models are read-only', [], 403);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Read-only');
  });

  it('maps known code "ru_overflow" to specific copy', () => {
    const err = new ApiError('ru_overflow', 'RU overflow', [], 400);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Rack full');
  });

  it('maps known code "position_overlap" to specific copy', () => {
    const err = new ApiError('position_overlap', 'position overlap', [], 400);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Position taken');
  });

  it('maps known code "conflict" to specific copy', () => {
    const err = new ApiError('conflict', 'conflict', [], 409);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Conflict');
  });

  it('maps known code "agg_ports_full" to specific copy', () => {
    const err = new ApiError('agg_ports_full', 'aggregation ports full', [], 422);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('No available ports');
  });

  it('maps known code "agg_model_downsize" to specific copy', () => {
    const err = new ApiError('agg_model_downsize', 'agg model downsize', [], 422);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Model downsize blocked');
  });

  it('falls back to server message for unknown ApiError code', () => {
    const err = new ApiError('some_future_code', 'Something the server said', [], 500);
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Error');
    expect(copy.description).toBe('Something the server said');
  });

  it('returns generic message for plain Error (network failure etc.)', () => {
    const err = new Error('Failed to fetch');
    const copy = errorToCopy(err);
    expect(copy.title).toBe('Something went wrong');
    expect(copy.description).toContain('unexpected error');
  });

  it('returns generic message for null', () => {
    const copy = errorToCopy(null);
    expect(copy.title).toBe('Something went wrong');
  });

  it('returns generic message for undefined', () => {
    const copy = errorToCopy(undefined);
    expect(copy.title).toBe('Something went wrong');
  });

  it('returns generic message for a string thrown', () => {
    const copy = errorToCopy('something bad');
    expect(copy.title).toBe('Something went wrong');
  });
});
