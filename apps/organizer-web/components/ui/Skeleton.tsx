import React from "react";

export interface SkeletonProps {
  className?: string;
}

export const Skeleton: React.FC<SkeletonProps> = ({ className = "" }) => {
  return (
    <div
      className={`animate-pulse bg-[#E4E7F0] rounded-lg ${className}`}
      aria-hidden="true"
    />
  );
};

export const EventCardSkeleton: React.FC = () => {
  return (
    <div className="bg-white border border-[#E4E7F0] rounded-xl overflow-hidden shadow-sm p-5 flex flex-col justify-between h-[280px]">
      <div>
        <div className="flex justify-between items-center mb-3">
          <Skeleton className="h-5 w-20 rounded-full" />
          <Skeleton className="h-4 w-24" />
        </div>
        <Skeleton className="h-6 w-3/4 mb-2" />
        <Skeleton className="h-4 w-full mb-1" />
        <Skeleton className="h-4 w-2/3 mb-4" />
        <Skeleton className="h-2 w-full rounded-full mb-4" />
      </div>
      <div className="flex justify-between items-center pt-3 border-t border-[#E4E7F0]">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-8 w-20 rounded-lg" />
      </div>
    </div>
  );
};
