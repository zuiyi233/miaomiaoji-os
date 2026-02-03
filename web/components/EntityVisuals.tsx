
import React from 'react';
import { EntityType } from '../types';
import { Users, Shield, Sword, BookOpen, Zap, Scroll } from 'lucide-react';

export const getTypeStyle = (type: EntityType) => {
  switch (type) {
    case 'character': return 'bg-rose-50 text-rose-600 border-rose-100';
    case 'organization': return 'bg-amber-50 text-amber-600 border-amber-100';
    case 'item': return 'bg-cyan-50 text-cyan-600 border-cyan-100';
    case 'setting': return 'bg-emerald-50 text-emerald-600 border-emerald-100';
    case 'magic': return 'bg-indigo-50 text-indigo-600 border-indigo-100';
    case 'event': return 'bg-orange-50 text-orange-600 border-orange-100';
    default: return 'bg-gray-50 text-gray-600 border-gray-100';
  }
};

export const getEntityIcon = (type: EntityType) => {
  switch (type) {
    case 'character': return <Users className="w-4 h-4" />;
    case 'organization': return <Shield className="w-4 h-4" />;
    case 'item': return <Sword className="w-4 h-4" />;
    case 'setting': return <BookOpen className="w-4 h-4" />;
    case 'magic': return <Zap className="w-4 h-4" />;
    case 'event': return <Scroll className="w-4 h-4" />;
    default: return <BookOpen className="w-4 h-4" />;
  }
};
