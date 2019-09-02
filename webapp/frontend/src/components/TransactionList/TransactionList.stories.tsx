import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { TransactionList } from '.';
import { MemoryRouter } from 'react-router-dom';
import { TransactionItem } from '../../dataObjects/item';

const stories = storiesOf('components/TransactionList', module);

const item: TransactionItem = {
  id: 1,
  status: 'trading',
  transactionEvidenceStatus: 'wait_shipping',
  shippingStatus: 'initial',
  name: 'テスト商品',
  thumbnailUrl: 'https://i.gyazo.com/8560fce19556b64c95ad091350910184.jpg',
  createdAt: 111111111,
};

const mockProps = {
  items: [item],
  hasNext: false,
  loadMore: (createdAt: number, itemId: number, page: number) => {},
};

/**
stories
  .add('default', () => (
    <MemoryRouter>
      <TransactionList {...mockProps} />
    </MemoryRouter>
  ))
  .add('many transactions', () => (
    <MemoryRouter>
      <TransactionList
        {...mockProps}
        items={Array(30)
          .fill(item)
          .map((item: TransactionItem, index: number) =>
            Object.assign({}, item, { id: item.id + index, onClickTransaction: (item: TransactionItem) => {}}),
          )}
      />
    </MemoryRouter>
  ));
 */
