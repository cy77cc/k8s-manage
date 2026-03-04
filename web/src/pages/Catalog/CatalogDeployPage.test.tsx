import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CatalogDeployPage from './CatalogDeployPage';

const mockGetTemplate = vi.fn();
const mockPreview = vi.fn();
const mockDeploy = vi.fn();

vi.mock('../../api', () => ({
  Api: {
    catalog: {
      getTemplate: (...args: unknown[]) => mockGetTemplate(...args),
      preview: (...args: unknown[]) => mockPreview(...args),
      deploy: (...args: unknown[]) => mockDeploy(...args),
    },
  },
}));

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...(actual as object),
    useNavigate: () => mockNavigate,
  };
});

describe('CatalogDeployPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetTemplate.mockResolvedValue({
      data: {
        id: 100,
        name: 'mysql',
        display_name: 'MySQL',
        compose_template: '',
        variables_schema: [{ name: 'root_password', type: 'password', required: true }],
      },
    });
    mockPreview.mockResolvedValue({ data: { rendered_yaml: 'kind: Deployment', unresolved_vars: [] } });
    mockDeploy.mockResolvedValue({ data: { service_id: 10, template_id: 100, deploy_count: 1 } });
  });

  it('renders dynamic variable input and previews yaml', async () => {
    render(
      <MemoryRouter initialEntries={['/catalog/100/deploy']}>
        <Routes>
          <Route path="/catalog/:id/deploy" element={<CatalogDeployPage />} />
        </Routes>
      </MemoryRouter>
    );

    await screen.findByText('模板部署');
    expect(screen.getByText('root_password (password)')).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('例如 1'), { target: { value: '1' } });
    fireEvent.change(screen.getByPlaceholderText('请输入服务名称'), { target: { value: 'mysql-a' } });
    fireEvent.change(screen.getByLabelText('root_password (password)'), { target: { value: 'secret' } });

    fireEvent.click(screen.getByText('预览 YAML'));
    expect(await screen.findByDisplayValue('kind: Deployment')).toBeInTheDocument();
  });
});
