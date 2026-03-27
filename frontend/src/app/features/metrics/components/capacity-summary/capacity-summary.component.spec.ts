import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { CapacitySummaryComponent } from './capacity-summary.component';
import { CapacitySummary } from '../../../../models';

const mockSummary: CapacitySummary = {
  level: 'design',
  id: 1,
  name: 'my-design',
  power_watts_idle: 1000,
  power_watts_typical: 3000,
  power_watts_max: 5000,
  total_vcpu: 256,
  total_ram_gb: 2048,
  total_storage_tb: 10,
  total_gpu_count: 0,
  device_count: 12,
};

describe('CapacitySummaryComponent', () => {
  let fixture: ComponentFixture<CapacitySummaryComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CapacitySummaryComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    fixture = TestBed.createComponent(CapacitySummaryComponent);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should display capacity summary after loading', () => {
    fixture.componentRef.setInput('designId', 1);
    fixture.componentRef.setInput('level', 'design');
    fixture.detectChanges();

    const req = httpMock.expectOne('/api/designs/1/capacity?level=design');
    req.flush(mockSummary);
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('my-design');
    expect(text).toContain('3,000');
  });

  it('should display error message on failure', () => {
    fixture.componentRef.setInput('designId', 1);
    fixture.componentRef.setInput('level', 'design');
    fixture.detectChanges();

    const req = httpMock.expectOne('/api/designs/1/capacity?level=design');
    req.error(new ProgressEvent('error'));
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Failed to load capacity data.');
  });

  it('should reload when designId changes', () => {
    fixture.componentRef.setInput('designId', 1);
    fixture.componentRef.setInput('level', 'design');
    fixture.detectChanges();

    const req1 = httpMock.expectOne('/api/designs/1/capacity?level=design');
    req1.flush(mockSummary);
    fixture.detectChanges();
    fixture.detectChanges();

    fixture.componentRef.setInput('designId', 2);
    fixture.detectChanges();

    const req2 = httpMock.expectOne('/api/designs/2/capacity?level=design');
    req2.flush({ ...mockSummary, id: 2, name: 'design-2' });
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('design-2');
  });
});
