import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { ItemList, Props } from '.';

const stories = storiesOf('components/ItemList', module);

const mockProps: Props = {
  items: [],
  hasNext: false,
  loadMore: page => {},
};

stories.add('no items', () => <ItemList {...mockProps} />);
