import React, { useEffect, useRef, useState, useMemo } from 'react';
import { Project, StoryEntity, Document, EntityType } from '../types';
// Fix: added Activity to imports
import { X, ZoomIn, ZoomOut, Move, Activity } from 'lucide-react';

interface Node {
  id: string;
  label: string;
  type: EntityType | 'document';
  x: number;
  y: number;
  vx: number;
  vy: number;
}

interface Edge {
  source: string;
  target: string;
  label: string;
}

interface GraphVisualizerProps {
  project: Project;
  onClose: () => void;
}

export const GraphVisualizer: React.FC<GraphVisualizerProps> = ({ project, onClose }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [zoom, setZoom] = useState(0.8);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragNode, setDragNode] = useState<Node | null>(null);
  const [lastMousePos, setLastMousePos] = useState({ x: 0, y: 0 });

  const nodes = useMemo(() => {
    const n: Node[] = [];
    project.entities.forEach(e => n.push({ id: e.id, label: e.title, type: e.type, x: Math.random() * 800, y: Math.random() * 600, vx: 0, vy: 0 }));
    project.documents.forEach(d => n.push({ id: d.id, label: d.title, type: 'document', x: Math.random() * 800, y: Math.random() * 600, vx: 0, vy: 0 }));
    return n;
  }, [project.entities, project.documents]);

  const edges = useMemo(() => {
    const e: Edge[] = [];
    project.entities.forEach(entity => entity.linkedIds.forEach(l => e.push({ source: entity.id, target: l.targetId, label: l.relationName })));
    project.documents.forEach(doc => doc.linkedIds.forEach(l => e.push({ source: doc.id, target: l.targetId, label: l.relationName })));
    return e;
  }, [project.entities, project.documents]);

  useEffect(() => {
    let animationFrameId: number;
    const simulation = () => {
      const k = 0.04;
      const length = 180;
      const repulsion = 2500;

      nodes.forEach((n1, i) => {
        if (n1 === dragNode) return;
        n1.vx += (400 - n1.x) * 0.0005;
        n1.vy += (300 - n1.y) * 0.0005;

        nodes.forEach((n2, j) => {
          if (i === j) return;
          const dx = n2.x - n1.x;
          const dy = n2.y - n1.y;
          const dist = Math.sqrt(dx * dx + dy * dy) || 1;
          const f = repulsion / (dist * dist);
          n1.vx -= (dx / dist) * f;
          n1.vy -= (dy / dist) * f;
        });
      });

      edges.forEach(edge => {
        const s = nodes.find(n => n.id === edge.source);
        const t = nodes.find(n => n.id === edge.target);
        if (s && t) {
          const dx = t.x - s.x;
          const dy = t.y - s.y;
          const dist = Math.sqrt(dx * dx + dy * dy) || 1;
          const f = (dist - length) * k;
          const fx = (dx / dist) * f;
          const fy = (dy / dist) * f;
          if (s !== dragNode) { s.vx += fx; s.vy += fy; }
          if (t !== dragNode) { t.vx -= fx; t.vy -= fy; }
        }
      });

      nodes.forEach(n => {
        n.x += n.vx; n.y += n.vy;
        n.vx *= 0.9; n.vy *= 0.9;
      });

      draw();
      animationFrameId = requestAnimationFrame(simulation);
    };

    const draw = () => {
      const canvas = canvasRef.current;
      if (!canvas) return;
      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.save();
      ctx.translate(canvas.width / 2 + offset.x, canvas.height / 2 + offset.y);
      ctx.scale(zoom, zoom);
      ctx.translate(-400, -300);

      ctx.strokeStyle = '#f1f1f1';
      ctx.lineWidth = 1;
      edges.forEach(edge => {
        const s = nodes.find(n => n.id === edge.source);
        const t = nodes.find(n => n.id === edge.target);
        if (s && t) {
          ctx.beginPath();
          ctx.moveTo(s.x, s.y);
          ctx.lineTo(t.x, t.y);
          ctx.stroke();
          ctx.fillStyle = '#cbd5e1';
          ctx.font = '600 9px Inter';
          ctx.textAlign = 'center';
          ctx.fillText(edge.label, (s.x + t.x) / 2, (s.y + t.y) / 2 - 5);
        }
      });

      nodes.forEach(n => {
        ctx.beginPath();
        ctx.arc(n.x, n.y, n.type === 'document' ? 35 : 28, 0, Math.PI * 2);
        
        const colors: Record<string, string> = {
          character: '#f43f5e',
          organization: '#f59e0b',
          item: '#06b6d4',
          setting: '#10b981',
          magic: '#6366f1',
          event: '#f97316',
          document: '#475569'
        };
        
        ctx.fillStyle = colors[n.type] || '#ccc';
        ctx.fill();
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 3;
        ctx.stroke();

        ctx.fillStyle = '#fff';
        ctx.font = 'black 11px Inter';
        ctx.textAlign = 'center';
        ctx.fillText(n.label.substring(0, 8), n.x, n.y + 4);
      });

      ctx.restore();
    };

    simulation();
    return () => cancelAnimationFrame(animationFrameId);
  }, [nodes, edges, zoom, offset, dragNode]);

  const handleMouseDown = (e: React.MouseEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mouseX = (e.clientX - rect.left - canvas.width / 2 - offset.x) / zoom + 400;
    const mouseY = (e.clientY - rect.top - canvas.height / 2 - offset.y) / zoom + 300;
    const hit = nodes.find(n => Math.sqrt((n.x - mouseX)**2 + (n.y - mouseY)**2) < 30);
    if (hit) setDragNode(hit);
    else { setIsDragging(true); setLastMousePos({ x: e.clientX, y: e.clientY }); }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (dragNode) {
      const canvas = canvasRef.current;
      if (!canvas) return;
      const rect = canvas.getBoundingClientRect();
      dragNode.x = (e.clientX - rect.left - canvas.width / 2 - offset.x) / zoom + 400;
      dragNode.y = (e.clientY - rect.top - canvas.height / 2 - offset.y) / zoom + 300;
    } else if (isDragging) {
      setOffset(prev => ({ x: prev.x + (e.clientX - lastMousePos.x), y: prev.y + (e.clientY - lastMousePos.y) }));
      setLastMousePos({ x: e.clientX, y: e.clientY });
    }
  };

  return (
    <div className="fixed inset-0 z-[100] bg-white/95 backdrop-blur-xl flex flex-col">
      <div className="h-16 border-b border-gray-100 flex items-center justify-between px-8 bg-white/50">
        <h2 className="text-xl font-black text-gray-900 tracking-tight flex items-center gap-3">
          <Activity className="w-5 h-5 text-brand-500" />
          叙事全要素脉络
        </h2>
        <div className="flex items-center gap-4">
          <div className="flex gap-2">
            {['角色', '组织', '物品', '地点', '力量', '事件', '章节'].map((t, i) => (
              <span key={t} className="text-[9px] font-black uppercase px-2 py-0.5 rounded-md border border-gray-100 text-gray-400">{t}</span>
            ))}
          </div>
          <button onClick={onClose} className="p-3 bg-gray-900 text-white rounded-2xl hover:bg-gray-800 transition-all"><X className="w-5 h-5" /></button>
        </div>
      </div>
      <div className="flex-1 relative cursor-grab active:cursor-grabbing">
        <canvas
          ref={canvasRef}
          width={window.innerWidth}
          height={window.innerHeight - 64}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={() => { setDragNode(null); setIsDragging(false); }}
          className="w-full h-full"
        />
      </div>
    </div>
  );
};