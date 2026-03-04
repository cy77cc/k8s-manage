import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Space, Table, Tag, message } from 'antd';
import { Api } from '../../api';
import type { CatalogCategory } from '../../api/modules/catalog';

const CategoryManagePage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<CatalogCategory[]>([]);
  const [editing, setEditing] = useState<CatalogCategory | null>(null);
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await Api.catalog.listCategories();
      const rows = (resp.data.list || []).slice().sort((a, b) => a.sort_order - b.sort_order);
      setList(rows);
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载分类失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setOpen(true);
  };

  const openEdit = (item: CatalogCategory) => {
    setEditing(item);
    form.setFieldsValue(item);
    setOpen(true);
  };

  const save = async () => {
    const values = await form.validateFields();
    try {
      if (editing) {
        await Api.catalog.updateCategory(editing.id, values);
      } else {
        await Api.catalog.createCategory(values);
      }
      message.success('保存成功');
      setOpen(false);
      void load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '保存失败');
    }
  };

  const remove = async (id: number) => {
    try {
      await Api.catalog.deleteCategory(id);
      message.success('删除成功');
      void load();
    } catch (err) {
      message.error(err instanceof Error ? err.message : '删除失败');
    }
  };

  return (
    <div className="p-6 space-y-4">
      <Card extra={<Button type="primary" onClick={openCreate}>创建分类</Button>} title="分类管理">
        <Table
          loading={loading}
          rowKey="id"
          dataSource={list}
          columns={[
            { title: '名称', dataIndex: 'display_name' },
            { title: '标识', dataIndex: 'name' },
            { title: '排序', dataIndex: 'sort_order' },
            { title: '图标', dataIndex: 'icon' },
            { title: '类型', dataIndex: 'is_system', render: (v: boolean) => (v ? <Tag>系统</Tag> : <Tag color="green">自定义</Tag>) },
            {
              title: '操作',
              render: (_, row: CatalogCategory) => (
                <Space>
                  <Button size="small" onClick={() => openEdit(row)}>编辑</Button>
                  <Popconfirm
                    title={row.is_system ? '系统分类不可删除' : '确认删除该分类？'}
                    disabled={row.is_system}
                    onConfirm={() => remove(row.id)}
                  >
                    <Button size="small" danger disabled={row.is_system}>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title={editing ? '编辑分类' : '创建分类'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={save}
      >
        <Form layout="vertical" form={form}>
          {!editing && (
            <Form.Item name="name" label="分类标识" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
          )}
          <Form.Item name="display_name" label="显示名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="icon" label="图标">
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber className="w-full" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CategoryManagePage;
