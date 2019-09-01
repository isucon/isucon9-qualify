import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { Header } from '.';

const stories = storiesOf('components/Header', module);

const mockProps = {
  isLoggedIn: false,
  ownUserId: 1111,
  categories: [
    {
      id: 2,
      parentId: 1,
      categoryName: 'カテゴリ1',
    },
    {
      id: 3,
      parentId: 1,
      categoryName: 'カテゴリ2',
    },
    {
      id: 4,
      parentId: 1,
      categoryName: 'カテゴリ3',
    },
  ],
  goToTopPage: () => {},
  goToUserPage: (userId: number) => {},
  goToSettingPage: () => {},
  goToCategoryItemList: (categoryId: number) => {},
  onClickTitle: (isLoggedIn: boolean) => {},
};

stories
  .add('non sign in', () => <Header {...mockProps} />)
  .add('signed in', () => <Header {...mockProps} isLoggedIn={true} />);
