export interface Category {
  id: number;
  parentId: number;
  categoryName: string;
  parentCategoryName: string;
}

export interface CategorySimple {
  id: number;
  parentId: number;
  categoryName: string;
}
