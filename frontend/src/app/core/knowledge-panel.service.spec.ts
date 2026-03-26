import { TestBed } from '@angular/core/testing';
import { KnowledgePanelService } from './knowledge-panel.service';

describe('KnowledgePanelService', () => {
  let service: KnowledgePanelService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(KnowledgePanelService);
  });

  it('should start with panel closed', () => {
    expect(service.state().isOpen).toBe(false);
    expect(service.state().articlePath).toBeNull();
  });

  it('should open with the given article path', () => {
    service.open('networking/clos');
    expect(service.state().isOpen).toBe(true);
    expect(service.state().articlePath).toBe('networking/clos');
  });

  it('should close the panel', () => {
    service.open('networking/clos');
    service.close();
    expect(service.state().isOpen).toBe(false);
    expect(service.state().articlePath).toBeNull();
  });

  it('should toggle closed when opening the same article twice', () => {
    service.open('networking/clos');
    service.open('networking/clos');
    expect(service.state().isOpen).toBe(false);
  });

  it('should switch articles without closing', () => {
    service.open('networking/clos');
    service.open('networking/ecmp');
    expect(service.state().isOpen).toBe(true);
    expect(service.state().articlePath).toBe('networking/ecmp');
  });
});
