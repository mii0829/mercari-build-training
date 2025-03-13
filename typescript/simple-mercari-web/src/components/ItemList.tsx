import { useEffect, useState } from 'react';
import { type Item, fetchItems } from '~/api';

const PLACEHOLDER_IMAGE = `${import.meta.env.VITE_FRONTEND_URL}/logo192.png`;

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
  items?: Item[];
  onCategoryClick?: (category: string) => void;
}

export const ItemList = ({ reload = false, onLoadCompleted, items, onCategoryClick }: Prop) => {
  const [itemList, setItemList] = useState<Item[]>([]);

  useEffect(() => {
    if (reload) {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data);
          setItemList(data.items);
          onLoadCompleted && onLoadCompleted();
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    }
  }, [reload, onLoadCompleted]);

  const displayItems = items || itemList;

  return (
    <div className='ItemListContainer'>
      {displayItems.length > 0 ? (
        displayItems.map((item) => {
          const imageUrl = item.image_name
            ? `${import.meta.env.VITE_BACKEND_URL}/images/${item.image_name}`
            : PLACEHOLDER_IMAGE;
          return (
            <div key={item.id} className="ItemList">
              <img src={imageUrl} alt={item.name} className='itemImage' />
              <p className='info'>
                <span>Name: {item.name}</span>
                <br />
                <span>Category: {item.category}</span>
              </p>
              <p
                className="tag"
                onClick={() => onCategoryClick && onCategoryClick(item.category)}
                style={{ cursor: 'pointer', color: '#E83D11', textDecoration: 'underline' }}
              >
                {item.category}
              </p>
            </div>
          );
        })
      ) : (
        <p>No items found.</p>
      )}
    </div>
  );
};
