import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CatalogListPage from './CatalogListPage';

const mockListCategories = vi.fn();
const mockListTemplates = vi.fn();

vi.mock('../../api', () => ({
  Api: {
    catalog: {
      listCategories: (...args: unknown[]) => mockListCategories(...args),
      listTemplates: (...args: unknown[]) => mockListTemplates(...args),
    },
  },
}));

describe('CatalogListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListCategories.mockResolvedValue({ data: { list: [{ id: 1, display_name: '数据库' }], total: 1 } });
    mockListTemplates.mockResolvedValue({
      data: {
        list: [
          { id: 11, display_name: 'MySQL', name: 'mysql', description: 'db', tags: ['mysql'], deploy_count: 3, category_id: 1 },
          { id: 12, display_name: 'Redis', name: 'redis', description: 'cache', tags: ['redis'], deploy_count: 1, category_id: 2 },
        ],
        total: 2,
      },
    });
  });

  it('filters templates by search keyword', async () => {
    render(
      <MemoryRouter initialEntries={['/catalog']}>
        <Routes>
          <Route path="/catalog" element={<CatalogListPage />} />
        </Routes>
      </MemoryRouter>
    );

    await screen.findByText('MySQL');
    fireEvent.change(screen.getByPlaceholderText('搜索模板名称、描述或标签'), { target: { value: 'redis' } });
    expect(screen.queryByText('MySQL')).toBeNull();
    expect(await screen.findByText('Redis')).toBeInTheDocument();
  });
});
