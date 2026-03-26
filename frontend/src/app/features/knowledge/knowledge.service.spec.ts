import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { HttpTestingController } from '@angular/common/http/testing';
import { KnowledgeService } from './knowledge.service';
import { KnowledgeArticle, KnowledgeIndex } from '../../models';

const MOCK_ARTICLES: KnowledgeArticle[] = [
  {
    path: 'networking/clos',
    title: 'Clos Fabric Fundamentals',
    category: 'networking',
    tags: ['clos', 'fabric', 'spine-leaf'],
  },
  {
    path: 'networking/ecmp',
    title: 'ECMP Load Balancing',
    category: 'networking',
    tags: ['ecmp', 'multipath'],
  },
  {
    path: 'infrastructure/rack',
    title: 'Rack Design Basics',
    category: 'infrastructure',
    tags: ['rack', 'physical'],
  },
];

describe('KnowledgeService', () => {
  let service: KnowledgeService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(KnowledgeService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('getIndex fetches from /api/knowledge', () => {
    let result: KnowledgeIndex | undefined;
    service.getIndex().subscribe(r => (result = r));

    const req = http.expectOne('/api/knowledge');
    expect(req.request.method).toBe('GET');
    req.flush({ articles: MOCK_ARTICLES });

    expect(result?.articles.length).toBe(3);
  });

  it('getArticle fetches from /api/knowledge/:path', () => {
    let result: KnowledgeArticle | undefined;
    service.getArticle('networking/clos').subscribe(r => (result = r));

    const req = http.expectOne('/api/knowledge/networking/clos');
    expect(req.request.method).toBe('GET');
    req.flush({ ...MOCK_ARTICLES[0], content: '# Clos Basics\n\nBody.' });

    expect(result?.content).toBeTruthy();
  });

  describe('search', () => {
    it('returns all articles for empty query', () => {
      const results = service.search(MOCK_ARTICLES, '');
      expect(results.length).toBe(3);
    });

    it('matches by title', () => {
      const results = service.search(MOCK_ARTICLES, 'clos');
      expect(results.length).toBeGreaterThan(0);
      expect(results[0].path).toBe('networking/clos');
    });

    it('matches by category', () => {
      const results = service.search(MOCK_ARTICLES, 'infrastructure');
      expect(results.length).toBe(1);
      expect(results[0].path).toBe('infrastructure/rack');
    });

    it('matches by tag', () => {
      const results = service.search(MOCK_ARTICLES, 'multipath');
      expect(results.length).toBe(1);
      expect(results[0].path).toBe('networking/ecmp');
    });

    it('returns empty for no matches', () => {
      const results = service.search(MOCK_ARTICLES, 'xyznosuchthing');
      expect(results.length).toBe(0);
    });
  });

  describe('groupByCategory', () => {
    it('groups articles by category', () => {
      const grouped = service.groupByCategory(MOCK_ARTICLES);
      expect(grouped.size).toBe(2);
      expect(grouped.get('networking')?.length).toBe(2);
      expect(grouped.get('infrastructure')?.length).toBe(1);
    });
  });
});
