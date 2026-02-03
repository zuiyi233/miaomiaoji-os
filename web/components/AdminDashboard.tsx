
import React, { useEffect } from 'react';
import { useProject } from '../contexts/ProjectContext';
import { ViewMode } from '../types';

/**
 * @deprecated 独立 AdminDashboard 已移除，功能集成到 UserSettings 'admin' 标签页。
 */
export const AdminDashboard: React.FC = () => {
  const { setViewMode } = useProject();
  
  useEffect(() => {
    // 自动重定向到设置中心的管理标签
    setViewMode(ViewMode.SETTINGS);
  }, []);

  return null;
};
