import { api } from './client';
import type { KnowledgeIndex, KnowledgeArticle } from '@/models';

export const knowledgeApi = {
  index: () => api.get<KnowledgeIndex>('/knowledge'),
  article: (path: string) => api.get<KnowledgeArticle>(`/knowledge/${path}`),
};
