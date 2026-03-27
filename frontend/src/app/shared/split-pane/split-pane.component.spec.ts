import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';

import { SplitPaneComponent } from './split-pane.component';

describe('SplitPaneComponent', () => {
  let component: SplitPaneComponent;
  let fixture: ComponentFixture<SplitPaneComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SplitPaneComponent],
      providers: [provideAnimations()],
    }).compileComponents();

    fixture = TestBed.createComponent(SplitPaneComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with default leftPct', () => {
    component.initialLeftPct = 55;
    fixture.detectChanges();
    // After ngAfterViewInit, leftPct should be set
    expect(component.leftPct()).toBeGreaterThan(0);
    expect(component.leftPct()).toBeLessThanOrEqual(100);
  });

  it('should clamp leftPct within min/max on resize', () => {
    component.minLeftPct = 20;
    component.maxLeftPct = 80;
    component.leftPct.set(95); // above max
    component.onWindowResize();
    expect(component.leftPct()).toBe(80);

    component.leftPct.set(10); // below min
    component.onWindowResize();
    expect(component.leftPct()).toBe(20);
  });

  it('should have split-container element', () => {
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.split-container')).toBeTruthy();
  });

  it('should have a split-divider element', () => {
    const el: HTMLElement = fixture.nativeElement;
    const divider = el.querySelector('.split-divider');
    expect(divider).toBeTruthy();
  });

  it('divider should have role separator', () => {
    const el: HTMLElement = fixture.nativeElement;
    const divider = el.querySelector('.split-divider');
    expect(divider?.getAttribute('role')).toBe('separator');
  });

  it('divider should be keyboard accessible (tabindex=0)', () => {
    const el: HTMLElement = fixture.nativeElement;
    const divider = el.querySelector('.split-divider');
    expect(divider?.getAttribute('tabindex')).toBe('0');
  });

  it('left arrow key decreases leftPct', () => {
    const initial = component.leftPct();
    component.onDividerKeydown(new KeyboardEvent('keydown', { key: 'ArrowLeft' }));
    expect(component.leftPct()).toBe(initial - 2);
  });

  it('right arrow key increases leftPct', () => {
    const initial = component.leftPct();
    component.onDividerKeydown(new KeyboardEvent('keydown', { key: 'ArrowRight' }));
    expect(component.leftPct()).toBe(initial + 2);
  });

  it('arrow key respects minimum constraint', () => {
    component.minLeftPct = 20;
    component.leftPct.set(21);
    component.onDividerKeydown(new KeyboardEvent('keydown', { key: 'ArrowLeft' }));
    expect(component.leftPct()).toBe(20);
  });

  it('arrow key respects maximum constraint', () => {
    component.maxLeftPct = 80;
    component.leftPct.set(79);
    component.onDividerKeydown(new KeyboardEvent('keydown', { key: 'ArrowRight' }));
    expect(component.leftPct()).toBe(80);
  });

  it('left and right pane elements exist', () => {
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.split-left')).toBeTruthy();
    expect(el.querySelector('.split-right')).toBeTruthy();
  });

  describe('session storage persistence', () => {
    afterEach(() => {
      sessionStorage.clear();
    });

    it('saves leftPct to sessionStorage on mouseup', () => {
      component.storageKey = 'test-pane';
      component.leftPct.set(65);
      component['_saveToSession'](65);
      expect(sessionStorage.getItem('fabrik-split-pane-test-pane')).toBe('65');
    });

    it('loads leftPct from sessionStorage on init', () => {
      component.storageKey = 'test-pane';
      component.minLeftPct = 20;
      component.maxLeftPct = 80;
      sessionStorage.setItem('fabrik-split-pane-test-pane', '70');
      const loaded = component['_loadFromSession']();
      expect(loaded).toBe(70);
    });

    it('returns null from sessionStorage when key is missing', () => {
      component.storageKey = 'nonexistent-key';
      const loaded = component['_loadFromSession']();
      expect(loaded).toBeNull();
    });

    it('clamps loaded value to min/max bounds', () => {
      component.storageKey = 'test-pane';
      component.minLeftPct = 20;
      component.maxLeftPct = 80;
      sessionStorage.setItem('fabrik-split-pane-test-pane', '95');
      const loaded = component['_loadFromSession']();
      expect(loaded).toBe(80);

      sessionStorage.setItem('fabrik-split-pane-test-pane', '5');
      const loaded2 = component['_loadFromSession']();
      expect(loaded2).toBe(20);
    });

    it('returns null when sessionStorage contains non-numeric value', () => {
      component.storageKey = 'test-pane';
      sessionStorage.setItem('fabrik-split-pane-test-pane', 'not-a-number');
      const loaded = component['_loadFromSession']();
      expect(loaded).toBeNull();
    });
  });
});
