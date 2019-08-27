import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemList, Props } from '.';
import { ItemStatus, TimelineItem } from '../../dataObjects/item';
import { MemoryRouter } from 'react-router-dom';

const stories = storiesOf('components/ItemList', module);

const getMockItem = (
  id: number,
  status: ItemStatus,
  name: string,
  price: number,
): TimelineItem => ({
  id,
  status,
  name,
  price,
  thumbnailUrl: 'https://i.gyazo.com/8560fce19556b64c95ad091350910184.jpg',
  createdAt: 11111,
});

const mockProps: Props = {
  items: [],
  hasNext: false,
  loadMore: page => {},
};

stories
  .add('no items', () => (
    <MemoryRouter>
      <ItemList {...mockProps} />
    </MemoryRouter>
  ))
  .add('items', () => {
    const mockItems: TimelineItem[] = [];
    for (let i = 1; i <= 30; i++) {
      mockItems.push(
        getMockItem(i, i % 2 ? 'on_sale' : 'sold_out', `test ${i}`, i * 1000),
      );
    }
    return (
      <MemoryRouter>
        <ItemList {...mockProps} items={mockItems} />
      </MemoryRouter>
    );
  });
